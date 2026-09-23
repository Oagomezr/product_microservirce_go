package mongorepo

import (
	"context"
	"product-microservice/internal/a_domain/model"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const errorDecodingProduct = "error decoding product"
const errorIteratingProducts = "error iterating products"

type ProductRepository struct {
	collection *mongo.Collection
}

func NewProductRepository(db *mongo.Database) *ProductRepository {
	return &ProductRepository{
		collection: db.Collection("products"),
	}
}

// Equivalent to my "FindProductAdapter.java"
// When writing this method, this struct already implements automatically the findPort(product_finder) interface
func (r *ProductRepository) Load(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	var entity ProductEntity

	// Searching using the BSON map with the ID field
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&entity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("product not found")
		}
		return nil, errors.Wrap(err, "error finding product by ID")
	}
	return entity.toDomain(), nil
}

// Equivalent to my "SearchProductAdapter.java"
// When writing this method, this struct already implements automatically the searchPort(product_searcher) interface
func (r *ProductRepository) Search(ctx context.Context, query string) ([]*model.Product, error) {
	filter := bson.M{
		"$or": []bson.M{
			{
				"name": bson.M{
					"$regex":   query,
					"$options": "i",
				},
			},
			{
				"description": bson.M{
					"$regex":   query,
					"$options": "i",
				},
			},
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "error searching products")
	}
	defer cursor.Close(ctx)

	var entities []*model.Product

	for cursor.Next(ctx) {
		var entity ProductEntity

		if err := cursor.Decode(&entity); err != nil {
			return nil, errors.Wrap(err, errorDecodingProduct)
		}

		entities = append(entities, entity.toDomain())
	}

	if err := cursor.Err(); err != nil {
		return nil, errors.Wrap(err, errorIteratingProducts)
	}

	return entities, nil
}

// Equivalent to my "ListProductsByCategoryAdapter.java"
// When writing this method, this struct already implements automatically the listProductsByCategoryPort(category_products_lister) interface
func (r *ProductRepository) LoadByCategory(ctx context.Context, categoryID uuid.UUID) ([]*model.Product, error) {
	filter := bson.M{"category_id": categoryID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "error finding products by category ID")
	}
	defer cursor.Close(ctx)

	var entities []*model.Product

	for cursor.Next(ctx) {
		var entity ProductEntity

		if err := cursor.Decode(&entity); err != nil {
			return nil, errors.Wrap(err, "error decoding product")
		}

		entities = append(entities, entity.toDomain())
	}

	if err := cursor.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating products")
	}
	return entities, nil
}

func (r *ProductRepository) Exist(ctx context.Context, id uuid.UUID) (bool, error) {
	filter := bson.M{"id": id}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, errors.Wrap(err, "error checking if product exists")
	}
	return count > 0, nil
}

func (r *ProductRepository) Upsert(ctx context.Context, product *model.Product) error {
	productEntity := fromDomain(product)
	filter := bson.M{"id": productEntity.ID}
	update := bson.M{"$set": productEntity}
	_, err := r.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return errors.Wrap(err, "error upserting product")
	}
	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return errors.Wrap(err, "error deleting product")
	}
	return nil
}
