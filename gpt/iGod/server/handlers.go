package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func speak(ctx context.Context, og *iGodService, in string) (string, error) {
	edict := og.edicts.Speak(og.values, in)
	return edict.String(), edict.Act(ctx)
}

func handleFormInput(c *gin.Context) {
	in, ok := c.GetPostForm("input")
	if !ok {
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("missing input text"))
		return
	}
	s, err := speak(c.Request.Context(), extractServiceDeity(c), in)
	if err != nil {
		_ = c.AbortWithError(http.StatusServiceUnavailable, errors.New("failed edict"))
		return
	}
	c.String(http.StatusOK, s)
}

func extractServiceDeity(c *gin.Context) *iGodService {
	name := ServiceDeity.String()
	if v := c.MustGet(name).(*iGodService); v != nil {
		return v
	}
	panic(errors.New("heretic!"))
}
