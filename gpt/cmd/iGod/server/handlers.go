package server

import (
		"errors"
		"github.com/gin-gonic/gin"
		"net/http"
)

func speak(god *iGodService, in string) (string, error) {
		out := god.edicts.Speak(god.values, in)
		return out.String(), out.Act()
}

func handleFormInput(c *gin.Context) {
		in, ok := c.GetPostForm("input")
		if !ok {
				_ = c.AbortWithError(http.StatusBadRequest, errors.New("missing input text"))
				return
		}
		s, err := speak(extractServiceDeity(c), in)
		if err != nil {
				_ = c.AbortWithError(http.StatusServiceUnavailable, errors.New("failed edict"))
				return
		}
		c.String(http.StatusOK, s)
}

func extractServiceDeity(c *gin.Context) *iGodService {
		if deity := c.MustGet(ServiceDeity).(*iGodService); deity != nil {
				return deity
		}
		panic(errors.New("missing deity"))
}

