package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"

	"gopkg.in/redis.v1"
)

type Item struct {
	Id   string `json:"id"`
	Text string `json:"text" binding:"required"`
}

func main() {
	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		IndentJSON: true,
		Layout:     "layout",
	}))
	client := redisClient()
	defer client.Close()
	m.Map(client)

	m.Get("/", Root)

	// Items
	m.Get("/items", GetItems)
	m.Post("/items", binding.Bind(Item{}), CreateItem)
	m.Delete("/items/:id", DeleteItem)

	m.Run()
}

func redisClient() *redis.Client {

	// try to use redis cluster
	if os.ExpandEnv("$TODO_REDIS_CLUSTER_SERVICE_HOST") != "" {
		client, err := redisSentinelClient()
		if err == nil {
			return client
		}
	}

	// try to use redis standalone
	if os.ExpandEnv("$TODO_REDIS_STANDALONE_SERVICE_HOST") != "" {
		client, err := redisStandaloneClient()
		if err == nil {
			return client
		}
	}

	// return default client
	return redis.NewTCPClient(&redis.Options{})
}

func redisSentinelClient() (*redis.Client, error) {
	host := os.ExpandEnv("$TODO_REDIS_CLUSTER_SERVICE_HOST")
	port := os.ExpandEnv("$TODO_REDIS_CLUSTER_SERVICE_PORT")
	addr := fmt.Sprintf("%s:%s", host, port)

	// check env for master name override
	var master string
	if os.ExpandEnv("$REDIS_MASTER_NAME") != "" {
		master = os.ExpandEnv("$REDIS_MASTER_NAME")
	} else {
		master = "mymaster"
	}

	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    master,
		SentinelAddrs: []string{addr},
	})
	status := client.Ping()
	_, err := status.Result()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func redisStandaloneClient() (*redis.Client, error) {
	host := os.ExpandEnv("$TODO_REDIS_STANDALONE_SERVICE_HOST")
	port := os.ExpandEnv("$TODO_REDIS_STANDALONE_SERVICE_PORT")
	addr := fmt.Sprintf("%s:%s", host, port)

	client := redis.NewTCPClient(&redis.Options{Addr: addr})
	status := client.Ping()
	_, err := status.Result()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func Root(r render.Render) {
	r.HTML(http.StatusOK, "index", nil)
}

func GetItems(r render.Render, c *redis.Client) {
	results, err := c.HGetAllMap("godo:items").Result()
	if err != nil {
		panic(err)
	}
	items := make([]Item, 0, len(results))
	for id, text := range results {
		items = append(items, Item{
			Id:   id,
			Text: text,
		})
	}
	r.JSON(http.StatusOK, items)
}

func CreateItem(r render.Render, c *redis.Client, item Item) {
	id, err := c.Incr("godo:itemId").Result()
	if err != nil {
		panic(err)
	}
	item.Id = strconv.FormatInt(id, 10)
	err = c.HSet("godo:items", item.Id, item.Text).Err()
	if err != nil {
		panic(err)
	}
	r.JSON(http.StatusOK, item)
}

func DeleteItem(r render.Render, params martini.Params, c *redis.Client) {
	err := c.HDel("godo:items", params["id"]).Err()
	if err != nil {
		panic(err)
	}
	r.JSON(http.StatusOK, nil)
}
