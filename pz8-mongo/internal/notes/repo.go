package notes

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrNotFound = errors.New("note not found")

type Repo struct {
	col *mongo.Collection
}

func NewRepo(db *mongo.Database) (*Repo, error) {
	col := db.Collection("notes")
	
	// Создаем текстовый индекс для полнотекстового поиска
	textIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "title", Value: "text"},
			{Key: "content", Value: "text"},
		},
	}
	
	// Создаем TTL индекс для автоудаления
	ttlIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "expiresAt", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}
	
	// Создаем индекс для уникальности заголовка
	uniqueIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "title", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	
	// Создаем все индексы
	_, err := col.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		textIndex,
		ttlIndex,
		uniqueIndex,
	})
	if err != nil { return nil, err }
	
	return &Repo{col: col}, nil
}

func (r *Repo) Create(ctx context.Context, title, content string, expiresAt *time.Time) (Note, error) {
	now := time.Now()
	n := Note{
		Title:     title, 
		Content:   content, 
		CreatedAt: now, 
		UpdatedAt: now,
		ExpiresAt: expiresAt,
	}
	res, err := r.col.InsertOne(ctx, n)
	if err != nil { return Note{}, err }
	n.ID = res.InsertedID.(primitive.ObjectID)
	return n, nil
}

func (r *Repo) ByID(ctx context.Context, idHex string) (Note, error) {
	oid, err := primitive.ObjectIDFromHex(idHex)
	if err != nil { return Note{}, ErrNotFound }
	var n Note
	if err := r.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&n); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) { return Note{}, ErrNotFound }
		return Note{}, err
	}
	return n, nil
}

func (r *Repo) List(ctx context.Context, q string, limit, skip int64) ([]Note, error) {
	filter := bson.M{}
	if q != "" {
		// Используем текстовый поиск вместо regex
		filter["$text"] = bson.M{"$search": q}
	}
	opts := options.Find().SetLimit(limit).SetSkip(skip).SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil { return nil, err }
	defer cur.Close(ctx)

	var out []Note
	for cur.Next(ctx) {
		var n Note
		if err := cur.Decode(&n); err != nil { return nil, err }
		out = append(out, n)
	}
	return out, cur.Err()
}

// ListCursor - пагинация по курсору (после указанного ID)
func (r *Repo) ListCursor(ctx context.Context, q string, limit int64, afterHex string) ([]Note, error) {
	filter := bson.M{}
	
	if afterHex != "" {
		afterID, err := primitive.ObjectIDFromHex(afterHex)
		if err != nil {
			// Если передан некорректный ID, игнорируем фильтр
		} else {
			filter["_id"] = bson.M{"$lt": afterID}
		}
	}
	
	if q != "" {
		filter["$text"] = bson.M{"$search": q}
	}
	
	// Сортировка по _id в обратном порядке для пагинации "назад"
	opts := options.Find().SetLimit(limit).SetSort(bson.D{{Key: "_id", Value: -1}})
	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil { return nil, err }
	defer cur.Close(ctx)

	var out []Note
	for cur.Next(ctx) {
		var n Note
		if err := cur.Decode(&n); err != nil { return nil, err }
		out = append(out, n)
	}
	return out, cur.Err()
}

func (r *Repo) Update(ctx context.Context, idHex string, title, content *string, expiresAt *time.Time) (Note, error) {
	oid, err := primitive.ObjectIDFromHex(idHex)
	if err != nil { return Note{}, ErrNotFound }

	set := bson.M{"updatedAt": time.Now()}
	if title != nil   { set["title"] = *title }
	if content != nil { set["content"] = *content }
	if expiresAt != nil { set["expiresAt"] = *expiresAt }

	after := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated Note
	if err := r.col.FindOneAndUpdate(ctx, bson.M{"_id": oid}, bson.M{"$set": set}, after).Decode(&updated); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) { return Note{}, ErrNotFound }
		return Note{}, err
	}
	return updated, nil
}

func (r *Repo) Delete(ctx context.Context, idHex string) error {
	oid, err := primitive.ObjectIDFromHex(idHex)
	if err != nil { return ErrNotFound }
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil { return err }
	if res.DeletedCount == 0 { return ErrNotFound }
	return nil
}

// GetStats - возвращает статистику заметок
func (r *Repo) GetStats(ctx context.Context) (map[string]interface{}, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.M{
			"_id": nil,
			"totalNotes": bson.M{"$sum": 1},
			"avgContentLength": bson.M{"$avg": bson.M{"$strLenCP": "$content"}},
		}}},

		{{Key: "$project", Value: bson.M{
			"_id": 0,
			"totalNotes": 1,
			"avgContentLength": bson.M{"$round": bson.A{"$avgContentLength", 2}},
		}}},
	}

	cur, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var result []map[string]interface{}
	if err = cur.All(ctx, &result); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return map[string]interface{}{
			"totalNotes": 0,
			"avgContentLength": 0,
		}, nil
	}

	return result[0], nil
}