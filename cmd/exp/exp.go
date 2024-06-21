package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var conn *pgxpool.Pool

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
}

func createTable() {
	query := `
	CREATE TABLE IF NOT EXISTS posts (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		content TEXT,
		author TEXT NOT NULL
	);
	`

	_, err := conn.Exec(context.Background(), query)
	if err != nil {
		panic(err)
	}

	fmt.Println("Table posts created")
}

func insertPostWithReturn() {
	title := "Post 3"
	content := "Conteúdo 3"
	author := "Rudson Alves"
	query := `
	INSERT INTO posts (title, content, author)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	row := conn.QueryRow(context.Background(), query, title, content, author)
	var id int
	err := row.Scan(&id)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Post id: %d\n", id)
}

func selectById(id int) {
	query := `
	SELECT title, content, author FROM posts WHERE id = $1
	`

	row := conn.QueryRow(context.Background(), query, id)
	var title, content, author string

	err := row.Scan(&title, &content, &author)
	// if err == pgxpool. .ErrNoRows {
	// 	fmt.Printf("No post find from id: %d\n", id)
	// 	return
	// } else
	if err != nil {
		panic(err)
	}
	fmt.Printf("title: %s, content = %s, author = %s\n", title, content, author)
}

type Post struct {
	Id      int
	Title   string
	Content string
	Author  string
}

func (p Post) String() string {
	return fmt.Sprintf("Post id: %d  title: '%s'  content: '%s'  author: '%s'",
		p.Id,
		p.Title,
		p.Content,
		p.Author,
	)
}

func selectAllPosts() (posts []Post) {
	query := `
	SELECT id, title, content, author FROM posts
	`

	rows, err := conn.Query(context.Background(), query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		panic(err)
	}

	for rows.Next() {
		var post Post
		err := rows.Scan(&post.Id, &post.Title, &post.Content, &post.Author)
		if err != nil {
			panic(err)
		}
		posts = append(posts, post)
	}

	return
}

func main() {
	loadEnv() // Loading .env variables
	var err error

	dbUrl := os.Getenv("DATABASE_URL")
	fmt.Printf("Connection Url: %s\n", dbUrl)
	conn, err = pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		panic(err)
	}
	defer conn.Close()

	createTable()
	// insertPostWithReturn()
	selectById(2)
	posts := selectAllPosts()

	for _, post := range posts {
		fmt.Println(post)
	}
}

// func main() {
// 	// h := slog.NewJSONHandler(os.Stdout, nil)
// 	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
// 		Level: slog.LevelDebug,
// 		// AddSource: true,
// 	})
// 	log := slog.New(h)

// 	log.Info("Info messages")
// 	log.Debug("Debug message", "request_id", 1, "user", "Rudson")
// 	log.Warn("Warn message")
// 	log.Error("Error message", "request_id", 1)
// }

// func main() {
// 	err := godotenv.Load()
// 	if err != nil {
// 		panic(err)
// 	}

// 	port := os.Getenv("SERVER_PORT")
// 	host := os.Getenv("HOST")

// 	fmt.Printf("Servidor está rodando na porta: %s:%s/", host, port)

// }

// func main() {
// 	var port int
// 	var host string
// 	var verbose bool

// 	flag.IntVar(&port, "port", 7000, "Server port")
// 	flag.StringVar(&host, "host", "suzail.com", "Hostname")
// 	flag.BoolVar(&verbose, "v", false, "Verbose mode")

// 	flag.Parse()

// 	if verbose {
// 		fmt.Printf("Server is running in: %s:%d\n", host, port)
// 	} else {
// 		fmt.Printf("%s:%d\n", host, port)
// 	}
// }

// type Config struct {
// 	Server struct {
// 		Port      int
// 		Host      string
// 		StaticDir string
// 	}
// }

// func main() {
// 	file, err := os.Open("config.json")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer file.Close()

// 	var config Config
// 	err = json.NewDecoder(file).Decode(&config)
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Printf("Static dir: %s\n", config.Server.StaticDir)
// 	fmt.Printf("Host: %s:%d\n", config.Server.Host, config.Server.Port)
// }
