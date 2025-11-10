package goresponse

import (
	"math"

	"github.com/gofiber/fiber/v3"
)

// HttpMetaResponse is the meta response for the http response
// it is used to provide pagination information
type HttpMetaResponse struct {
	Page      int `json:"page"`
	PerPage   int `json:"perPage"`
	Total     int `json:"total"`
	TotalPage int `json:"totalPage"`
}

// HttpResponse is the response for the http response
// it is used to provide the response data
type HttpResponse struct {
	Message string            `json:"message"`
	Data    any               `json:"data"`
	Meta    *HttpMetaResponse `json:"meta,omitempty"`
}

// buildMeta is a helper function to build the meta response
// it is used to provide pagination information
func buildMeta(
	page int,
	perPage int,
	total int,
) *HttpMetaResponse {
	totalPage := 1
	if total > 0 {
		totalPage = int(math.Ceil(float64(total) / float64(perPage)))
	}
	return &HttpMetaResponse{
		Page:      page,
		PerPage:   perPage,
		Total:     total,
		TotalPage: totalPage,
	}
}

// compose is a helper function to compose the http response
// it is used to provide the response data
func compose(
	message string,
	data any,
	meta *HttpMetaResponse,
) HttpResponse {
	return HttpResponse{
		Message: message,
		Data:    data,
		Meta:    meta,
	}
}

// res is the implementation of the GoResponseClient interface
type res struct {
}

// GoResponseClient is the interface for the GoResponseClient
// it is used to provide the response data
type GoResponseClient interface {
	Jsonify(
		c fiber.Ctx,
		status int,
		message string,
		data any,
		meta *HttpMetaResponse,
	) error
	CreateMeta(
		page int,
		perPage int,
		total int,
	) *HttpMetaResponse
}

// NewGoResponseClient is a helper function to create a new GoResponseClient
// it is used to create a new GoResponseClient
func NewGoResponseClient() GoResponseClient {
	return &res{}
}

// Jsonify is a helper function to jsonify the response
// it is used to provide the response data
func (r *res) Jsonify(
	c fiber.Ctx,
	status int,
	message string,
	data any,
	meta *HttpMetaResponse,
) error {
	return c.
		Status(status).
		JSON(compose(message, data, meta))
}

// CreateMeta is a helper function to create the meta response
// it is used to provide pagination information
func (r *res) CreateMeta(
	page int,
	perPage int,
	total int,
) *HttpMetaResponse {
	return buildMeta(page, perPage, total)
}
