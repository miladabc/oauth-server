package auth

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/milad-abbasi/oauth-server/pkg/common"
	"github.com/milad-abbasi/oauth-server/pkg/user"
	"go.uber.org/zap"
)

type Controller struct {
	l *zap.Logger
	r *echo.Echo
	v *common.Validator
	s *Service
}

func NewController(
	logger *zap.Logger,
	router *echo.Echo,
	validator *common.Validator,
	service *Service,
) *Controller {
	return &Controller{
		l: logger.Named("AuthController"),
		r: router,
		v: validator,
		s: service,
	}
}

func (con *Controller) RegisterRoutes() {
	authRouter := con.r.Group("/auth")
	authRouter.POST("/register", con.Register)
	authRouter.POST("/login", con.Login)
}

func (con *Controller) Register(c echo.Context) error {
	var dto RegisterDto
	if err := c.Bind(&dto); err != nil {
		return err
	}

	err := con.v.Validate(&dto)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	tokens, err := con.s.Register(
		c.Request().Context(),
		&RegisterInfo{
			Name:     dto.Name,
			Email:    dto.Email,
			Password: dto.Password,
		},
	)
	if err != nil {
		if errors.Is(err, user.ErrUserExists) {
			return echo.NewHTTPError(http.StatusConflict, err)
		}

		return err
	}

	return c.JSON(http.StatusOK, tokens)
}

func (con *Controller) Login(c echo.Context) error {
	var dto LoginDto
	if err := c.Bind(&dto); err != nil {
		return err
	}

	err := con.v.Validate(&dto)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	tokens, err := con.s.Login(
		c.Request().Context(),
		&LoginInfo{
			Email:    dto.Email,
			Password: dto.Password,
		},
	)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return echo.NewHTTPError(http.StatusUnauthorized, err)
		}

		return err
	}

	return c.JSON(http.StatusOK, tokens)
}
