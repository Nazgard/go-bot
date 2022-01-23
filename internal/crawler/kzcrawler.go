package crawler

import (
	"bytes"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/repository"
	"makarov.dev/bot/internal/service/kinozal"
	kinozalClient "makarov.dev/bot/pkg/kinozal"
	"strconv"
	"time"
)

type KinozalCrawler struct {
	Service kinozal.Service
	Client  *kinozalClient.Client
	Bucket  *gridfs.Bucket
}

func NewKinozalCrawler(service kinozal.Service, client *kinozalClient.Client, bucket *gridfs.Bucket) *KinozalCrawler {
	return &KinozalCrawler{
		Service: service,
		Client:  client,
		Bucket:  bucket,
	}
}

func (c *KinozalCrawler) Start() {
	log := config.GetLogger()
	if !config.GetConfig().Kinozal.Enable {
		log.Info("Kinozal integration disabled")
		return
	}
	ch := make(chan int64)
	go c.Client.Listing(ch, time.Minute)
	for id := range ch {
		favorite, err := c.Service.IsFavorite(id)
		if err != nil {
			log.Error(err.Error())
		}
		if !favorite {
			continue
		}

		name, err := c.Client.GetName(id)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		exist, err := c.Service.Exists(id, name)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		if exist {
			continue
		}

		element, _ := c.Client.GetElement(id)

		objectID, err := c.Bucket.UploadFromStream(strconv.FormatInt(id, 10)+".torrent", bytes.NewReader(element.Torrent))
		if err != nil {
			log.Error("Error while store torrent", err.Error())
			continue
		}

		err = c.Service.Save(&repository.KinozalItem{
			Id:       primitive.NewObjectID(),
			Name:     element.Name,
			DetailId: id,
			GridFsId: objectID,
			Created:  time.Now(),
		})
		if err != nil {
			log.Error(err.Error())
			continue
		}
	}
}
