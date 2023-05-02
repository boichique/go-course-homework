package main

import (
	"net/http"

	"github.com/cloudmachinery/apps/http-userroles/contracts"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{
		store: store,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	usersGroup := e.Group("/api/users")
	usersGroup.GET("", h.handleGetAllUsers)
	usersGroup.GET("/:email", h.handleGetUserByEmail)
	usersGroup.GET("/roles/:role", h.handleGetUsersByRole)
	usersGroup.POST("", h.handleCreateUser)
	usersGroup.PUT("", h.handleUpdateUser)
	usersGroup.DELETE("/:email", h.handleDeleteUser)
}

func (h *Handler) handleGetAllUsers(c echo.Context) error {
	users, err := h.store.GetUsers()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, users)
}

func (h *Handler) handleGetUserByEmail(c echo.Context) error {
	email := c.Param("email")

	if email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}

	user, err := h.store.GetUser(email)
	if err != nil {
		if err == ErrUserNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, user)
}

func (h *Handler) handleGetUsersByRole(c echo.Context) error {
	role := c.Param("role")
	if role == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "role is required")
	}

	users, err := h.store.GetUsersByRole(role)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, users)
}

func (h *Handler) handleCreateUser(c echo.Context) error {
	var user *contracts.User

	if err := c.Bind(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if user.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}

	err := h.store.CreateUser(user)

	switch err {
	case ErrUserAlreadyExists:
		return echo.NewHTTPError(http.StatusConflict, "user already exists")
	case nil:
		return c.NoContent(http.StatusCreated)
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
	}
}

func (h *Handler) handleUpdateUser(c echo.Context) error {
	var user *contracts.User

	if user.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}

	if err := c.Bind(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	err := h.store.UpdateUser(user)

	switch err {
	case ErrUserNotFound:
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	case nil:
		return c.NoContent(http.StatusOK)
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
}

func (h *Handler) handleDeleteUser(c echo.Context) error {
	email := c.Param("email")

	if email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}

	err := h.store.DeleteUser(email)

	switch err {
	case ErrUserNotFound:
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	case nil:
		return c.NoContent(http.StatusOK)
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
}
