package web

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"makarov.dev/bot/internal/service"
	"net/http"
)

func addFile(group *gin.RouterGroup, fileService service.FileService) {
	group.GET(":fileId", func(ctx *gin.Context) {
		fileId := ctx.Param("fileId")
		if fileId == "" {
			ctx.AbortWithStatus(400)
		}

		objectID, err := primitive.ObjectIDFromHex(fileId)
		if err != nil {
			ctx.AbortWithStatus(400)
		}

		reader, err := fileService.GetFile(&objectID)
		if err != nil {
			_ = ctx.AbortWithError(500, err)
		}
		file := reader.GetFile()
		extraHeaders := map[string]string{
			"Content-Disposition": "attachment; filename=" + file.Name,
		}
		ctx.DataFromReader(http.StatusOK, file.Length, file.Name, reader, extraHeaders)
	})
}
