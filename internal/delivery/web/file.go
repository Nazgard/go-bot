package web

import (
	"errors"
	"fmt"
	"makarov.dev/bot/internal/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"makarov.dev/bot/internal/service"
)

type FileController struct {
	FileService service.FileService
}

func (c *FileController) Add(g *gin.RouterGroup) {
	g.GET(":fileId", c.downloadFile())
}

// @Tags File controller
// @Param fileId path string true "File id"
// @Produce octet-stream
// @Produce json
// @Success 200 {file} file
// @Failure 400,500 {object} HTTPError
// @Router /dl/{fileId} [get]
func (c *FileController) downloadFile() func(ctx *gin.Context) {
	log := config.GetLogger()
	return func(ctx *gin.Context) {
		fileId := ctx.Param("fileId")
		if fileId == "" {
			NewError(ctx, 400, errors.New("bad objectId"))
			return
		}

		objectID, err := primitive.ObjectIDFromHex(fileId)
		if err != nil {
			NewError(ctx, 400, err)
			return
		}

		reader, err := c.FileService.GetFile(&objectID)
		if err != nil {
			NewError(ctx, 500, err)
			return
		}
		go func() {
			err = c.FileService.LogDownload(ctx, objectID)
			if err != nil {
				log.Error(fmt.Sprintf("Error while log download fileId=%s", objectID), err.Error())
			}
		}()
		file := reader.GetFile()
		extraHeaders := map[string]string{
			"Content-Disposition": "attachment; filename=" + file.Name,
		}
		ctx.DataFromReader(http.StatusOK, file.Length, file.Name, reader, extraHeaders)
	}
}
