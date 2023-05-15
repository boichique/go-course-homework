package main

import (
	"net/http"
	"net/url"

	"github.com/cloudmachinery/apps/http-userroles/contracts"
	_ "github.com/cloudmachinery/apps/http-userroles/server/docs"
	"github.com/cloudmachinery/apps/http-userroles/server/store"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	store store.Store
}

func NewHandler(store store.Store) *Handler {
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

//	@Summary		Returns all users
//	@Description	Returns all users
//	@Router			/api/users [get]
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	contracts.User
//	@Failure		500	{object}	echo.HTTPError
func (h *Handler) handleGetAllUsers(c echo.Context) error {
	users, err := h.store.GetUsers(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, users)
}

//	@Summary		Returns a user by email
//	@Description	Returns a user by email
//	@Router			/api/users/{email} [get]
//	@Accept			json
//	@Produce		json
//	@Param			email	path		string	true	"Email of the user"
//	@Success		200		{object}	contracts.User
//	@Failure		404		{object}	echo.HTTPError
//	@Failure		500		{object}	echo.HTTPError
func (h *Handler) handleGetUserByEmail(c echo.Context) error {
	email := c.Param("email")
	email, err := url.QueryUnescape(email)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}

	user, err := h.store.GetUser(c.Request().Context(), email)
	if err != nil {
		if err == store.ErrUserNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, user)
}

//	@Summary		Returns all users with the given role
//	@Description	Returns all users with the given role
//	@Router			/api/users/roles/{role} [get]
//	@Accept			json
//	@Produce		json
//	@Param			role	path		string	true	"Role of the user"
//	@Success		200		{array}		contracts.User
//	@Failure		400		{object}	echo.HTTPError
//	@Failure		500		{object}	echo.HTTPError
func (h *Handler) handleGetUsersByRole(c echo.Context) error {
	role := c.Param("role")
	if role == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "role is required")
	}

	users, err := h.store.GetUsersByRole(c.Request().Context(), role)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, users)
}

//	@Summary		Creates a new user
//	@Description	Creates a new user
//	@Router			/api/users [post]
//	@Accept			json
//	@Produce		json
//	@Param			user	body	contracts.User	true	"User object that needs to be created"
//	@Success		201
//	@Failure		400	{object}	echo.HTTPError
//	@Failure		409	{object}	echo.HTTPError
//	@Failure		500	{object}	echo.HTTPError
func (h *Handler) handleCreateUser(c echo.Context) error {
	var user *contracts.User

	if err := c.Bind(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if user.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}

	err := h.store.CreateUser(c.Request().Context(), user)

	switch err {
	case store.ErrUserAlreadyExists:
		return echo.NewHTTPError(http.StatusConflict, "user already exists")
	case nil:
		return c.NoContent(http.StatusCreated)
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
	}
}

//	@Summary		Updates an existing user
//	@Description	Updates an existing user
//	@Router			/api/users [put]
//	@Accept			json
//	@Produce		json
//	@Param			user	body	contracts.User	true	"User object that needs to be updated"
//	@Success		200
//	@Failure		400	{object}	echo.HTTPError
//	@Failure		404	{object}	echo.HTTPError
//	@Failure		500	{object}	echo.HTTPError
func (h *Handler) handleUpdateUser(c echo.Context) error {
	var user *contracts.User

	if err := c.Bind(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if user.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}

	err := h.store.UpdateUser(c.Request().Context(), user)

	switch err {
	case store.ErrUserNotFound:
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	case nil:
		return c.NoContent(http.StatusOK)
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
}

//	@Summary		Deletes a user by email
//	@Description	Deletes a user by email
//	@Router			/api/users/{email} [delete]
//	@Accept			json
//	@Produce		json
//	@Param			email	path	string	true	"Email of the user"
//	@Success		200
//	@Failure		400	{object}	echo.HTTPError
//	@Failure		404	{object}	echo.HTTPError
//	@Failure		500	{object}	echo.HTTPError
func (h *Handler) handleDeleteUser(c echo.Context) error {
	email := c.Param("email")
	email, err := url.QueryUnescape(email)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}

	err = h.store.DeleteUser(c.Request().Context(), email)

	switch err {
	case store.ErrUserNotFound:
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	case nil:
		return c.NoContent(http.StatusOK)
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
}
