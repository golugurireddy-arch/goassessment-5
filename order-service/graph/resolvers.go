package graph

import (
	"fmt"
	"order-service/model"
	"order-service/service"

	"github.com/graphql-go/graphql"
)

var orderStatusEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "OrderStatus",
	Values: graphql.EnumValueConfigMap{
		"CREATED":   &graphql.EnumValueConfig{Value: model.StatusCreated},
		"CONFIRMED": &graphql.EnumValueConfig{Value: model.StatusConfirmed},
		"CANCELLED": &graphql.EnumValueConfig{Value: model.StatusCancelled},
	},
})

var orderType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Order",
	Fields: graphql.Fields{
		"id":          &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"customerId":  &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"productId":   &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"quantity":    &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
		"totalAmount": &graphql.Field{Type: graphql.NewNonNull(graphql.Float)},
		"status":      &graphql.Field{Type: graphql.NewNonNull(orderStatusEnum)},
		"createdAt":   &graphql.Field{Type: graphql.String},
		"updatedAt":   &graphql.Field{Type: graphql.String},
	},
})

func BuildSchema(svc service.OrderService) (graphql.Schema, error) {
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"getOrder": &graphql.Field{
				Type: orderType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id := p.Args["id"].(string)
					return svc.GetOrder(p.Context, id)
				},
			},
			"orders": &graphql.Field{
				Type: graphql.NewList(orderType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return svc.ListOrders(p.Context)
				},
			},
		},
	})

	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createOrder": &graphql.Field{
				Type: orderType,
				Args: graphql.FieldConfigArgument{
					"customerId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"productId":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"quantity":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					customerID := p.Args["customerId"].(string)
					productID := p.Args["productId"].(string)
					quantity := p.Args["quantity"].(int)
					return svc.CreateOrder(p.Context, customerID, productID, quantity)
				},
			},
			"updateOrderStatus": &graphql.Field{
				Type: orderType,
				Args: graphql.FieldConfigArgument{
					"id":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"status": &graphql.ArgumentConfig{Type: graphql.NewNonNull(orderStatusEnum)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id := p.Args["id"].(string)
					statusVal, ok := p.Args["status"].(model.OrderStatus)
					if !ok {
						return nil, fmt.Errorf("invalid status value")
					}
					return svc.UpdateOrderStatus(p.Context, id, statusVal)
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})
}