package handler

import (
	"errors"
	"io"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gsxhnd/jaha/server/db/database"
	"github.com/gsxhnd/jaha/server/errno"
	"github.com/gsxhnd/jaha/server/model"
	"github.com/gsxhnd/jaha/server/service"
	"github.com/gsxhnd/jaha/server/storage"
	"github.com/gsxhnd/jaha/utils"
)

type MovieHandler interface {
	CreateMovies(ctx *fiber.Ctx) error
	DeleteMovies(ctx *fiber.Ctx) error
	UpdateMovie(ctx *fiber.Ctx) error
	UploadCover(ctx *fiber.Ctx) error
	GetMovies(ctx *fiber.Ctx) error
	GetMovieInfo(ctx *fiber.Ctx) error
	SearchMovies(ctx *fiber.Ctx) error
}

type movieHandle struct {
	valid   *validator.Validate
	svc     service.MovieService
	logger  utils.Logger
	storage storage.Storage
}

func NewMovieHandler(svc service.MovieService, v *validator.Validate, s storage.Storage, l utils.Logger) MovieHandler {
	return &movieHandle{
		valid:   v,
		svc:     svc,
		logger:  l,
		storage: s,
	}
}

// @Summary      Create movies
// @Description  Create movies
// @Tags         movie
// @Produce      json
// @Success      200  {object}  errno.errno
// @Router       /movie [post]
func (h *movieHandle) CreateMovies(ctx *fiber.Ctx) error {
	var body = make([]model.Movie, 0)
	if err := ctx.BodyParser(&body); err != nil {
		h.logger.Errorf(err.Error())
		return ctx.JSON(errno.DecodeError(err))
	}

	if err := h.valid.Var(body, "dive"); err != nil {
		return ctx.JSON(errno.DecodeError(err))
	}

	err := h.svc.CreateMovies(body)
	return ctx.JSON(errno.DecodeError(err))
}

// @Summary      Delete movies
// @Description  Delete movies
// @Tags         movie
// @Accept       json
// @Produce      json
// @Param        default body []uint true "default"
// @Success      200 {object} errno.errno{data=nil}
// @Router       /movie [delete]
func (h *movieHandle) DeleteMovies(ctx *fiber.Ctx) error {
	var body = make([]uint, 0)
	if err := ctx.BodyParser(&body); err != nil {
		h.logger.Errorf(err.Error())
		return ctx.JSON(errno.DecodeError(err))
	}

	if err := h.valid.Var(body, "dive"); err != nil {
		return ctx.JSON(errno.DecodeError(err))
	}

	err := h.svc.DeleteMovies(body)
	return ctx.JSON(errno.DecodeError(err))
}

// UpdateMovie implements MovieHandler.
// @Summary      Update a movie by id
// @Description  Update a movie by id
// @Tags         movie
// @Accept       json
// @Produce      json
// @Param        tag body model.Movie true "movie object"
// @Success      200 {object} errno.errno
// @Router       /movie [put]
func (h *movieHandle) UpdateMovie(ctx *fiber.Ctx) error {
	var body model.Movie
	if err := ctx.BodyParser(&body); err != nil {
		h.logger.Errorf(err.Error())
		return ctx.JSON(errno.DecodeError(err))
	}

	if err := h.valid.Struct(body); err != nil {
		h.logger.Errorf(err.Error())
		return ctx.JSON(errno.DecodeError(err))
	}

	if err := h.svc.UpdateMovie(&body); err != nil {
		h.logger.Errorf(err.Error())
		return ctx.JSON(errno.DecodeError(err))
	}

	return ctx.JSON(errno.OK)
}

// @Summary      Get movies
// @Description  Get movies
// @Tags         movie
// @Produce      json
// @Param        page_size query int false "int valid" default(50)
// @Param        page query int false "int valid" default(1)
// @Success      200 {object} errno.errno{data=[]model.Movie}
// @Router       /movie [get]
func (h *movieHandle) GetMovies(ctx *fiber.Ctx) error {
	var p = database.Pagination{
		Limit:  uint64(ctx.QueryInt("page_size", 50)),
		Offset: uint64(ctx.QueryInt("page_size", 50) * (ctx.QueryInt("page", 1) - 1)),
	}

	if err := h.valid.Struct(p); err != nil {
		return ctx.JSON(errno.DecodeError(err))
	}

	data, err := h.svc.GetMovies(&p)

	return ctx.JSON(errno.DecodeError(err).WithData(data))
}

// @Summary      Get movies
// @Description  Get movies
// @Tags         movie
// @Produce      json
// @Param        code path string true "movie code"
// @Success      200 {object} errno.errno{data=model.MovieInfo}
// @Router       /movie/info/:code [get]
func (h *movieHandle) GetMovieInfo(ctx *fiber.Ctx) error {
	code := ctx.Params("code", "")
	data, err := h.svc.GetMovieInfo(code)
	return ctx.JSON(errno.DecodeError(err).WithData(data))
}

// @Summary      Search movies
// @Description  Search movies by code
// @Tags         movie
// @Produce      json
// @Param        code query string true "movie code"
// @Success      200 {object} errno.errno{data=[]model.Movie}
// @Router       /movie/search [get]
func (h *movieHandle) SearchMovies(ctx *fiber.Ctx) error {
	data, err := h.svc.SearchMoviesByCode(ctx.Query("code"))
	if err != nil {
		return ctx.JSON(errno.DecodeError(err))
	}
	return ctx.JSON(errno.OK.WithData(data))
}

// @Summary      Upload movie cover
// @Description  Upload movie cover by movie id
// @Tags         movie
// @Produce      json
// @Param        code query string true "movie code"
// @Success      200 {object} errno.errno{data=[]model.Movie}
// @Router       /movie/cover [put]
func (h movieHandle) UploadCover(ctx *fiber.Ctx) error {
	code := ctx.Params("code", "")

	form, err := ctx.MultipartForm()
	if err != nil {
		return ctx.JSON(errno.DecodeError(err))
	}

	files := form.File["cover"]
	if files == nil || len(files) <= 0 || len(files) > 1 {
		return ctx.JSON(errno.DecodeError(errors.New("file is not exist or too many")))
	}

	file, err := files[0].Open()
	if err != nil {
		return ctx.JSON(errno.DecodeError(err))
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return ctx.JSON(errno.DecodeError(err))
	}

	err = h.svc.UploadMovieCover(code, files[0].Filename, fileBytes)
	return ctx.JSON(errno.DecodeError(err))
}
