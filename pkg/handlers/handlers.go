package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/j18e/backerupper/zsh"
)

func GetLastCMD(s zsh.Service) func(*gin.Context) {
	return func(ctx *gin.Context) {
		c, err := s.LastCMD()
		if err != nil {
			ctx.String(http.StatusInternalServerError, "internal server error")
			fmt.Fprintf(os.Stderr, "ERROR zsh.Service.LastCMD: %v\n", err)
			return
		}

		ctx.JSON(http.StatusOK, c)
	}
}

func WriteCMDs(s zsh.Service) func(*gin.Context) {
	return func(ctx *gin.Context) {
		var cx []zsh.Command

		if err := ctx.ShouldBindJSON(&cx); err != nil {
			resp := fmt.Sprintf("bad request: %v", err)
			ctx.String(http.StatusBadRequest, resp)
			return
		}
		if err := s.WriteCMDs(cx); err != nil {
			ctx.String(http.StatusInternalServerError, "internal server error")
			fmt.Fprintf(os.Stderr, "ERROR zsh.Service.WriteCMDs: %v\n", err)
			return
		}
		ctx.String(http.StatusOK, "done")
	}
}
