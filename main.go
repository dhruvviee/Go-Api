package main

import (
    "database/sql"
    "fmt"
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    _ "github.com/mattn/go-sqlite3"
)

type Task struct {
    ID          int       `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    DueDate     time.Time `json:"due_date"`
    Status      string    `json:"status"`
}

var db *sql.DB

func main() {
    // Initialize the database and set up routes
    initDB()

    router := gin.Default()

    // Add a handler for the root path
    router.GET("/", welcome)

    router.POST("/tasks", createTask)
    router.GET("/tasks/:id", getTask)
    router.PUT("/tasks/:id", updateTask)
    router.DELETE("/tasks/:id", deleteTask)
    router.GET("/tasks", listTasks)

    router.Run(":8080")
}

func initDB() {
    database, err := sql.Open("sqlite3", "./tass.db")
    if err != nil {
        fmt.Println("Error opening database:", err)
        return
    }
    db = database

    createTableSQL := `
    CREATE TABLE IF NOT EXISTS tasks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        description TEXT,
        due_date DATE,
        status TEXT
    );`
    _, err = db.Exec(createTableSQL)
    if err != nil {
        fmt.Println("Error creating tasks table:", err)
        return
    }
}

// Create a new task
func createTask(c *gin.Context) {
    var task Task
    if err := c.ShouldBindJSON(&task); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Insert task into the database
    stmt, err := db.Prepare("INSERT INTO tasks (title, description, due_date, status) VALUES (?, ?, ?, ?)")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
        return
    }
    defer stmt.Close()

    result, err := stmt.Exec(task.Title, task.Description, task.DueDate, task.Status)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
        return
    }

    id, _ := result.LastInsertId()
    task.ID = int(id)

    c.JSON(http.StatusCreated, task)
}

// Retrieve a task by ID
func getTask(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
        return
    }

    var task Task
    row := db.QueryRow("SELECT id, title, description, due_date, status FROM tasks WHERE id = ?", id)
    err = row.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
        return
    }

    c.JSON(http.StatusOK, task)
}

// Update a task by ID
func updateTask(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
        return
    }

    var task Task
    if err := c.ShouldBindJSON(&task); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Update task in the database
    stmt, err := db.Prepare("UPDATE tasks SET title = ?, description = ?, due_date = ?, status = ? WHERE id = ?")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
        return
    }
    defer stmt.Close()

    _, err = stmt.Exec(task.Title, task.Description, task.DueDate, task.Status, id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
        return
    }

    task.ID = id
    c.JSON(http.StatusOK, task)
}

// Delete a task by ID
func deleteTask(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
        return
    }

    // Delete task from the database
    _, err = db.Exec("DELETE FROM tasks WHERE id = ?", id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// List all tasks
func listTasks(c *gin.Context) {
    var tasks []Task

    rows, err := db.Query("SELECT id, title, description, due_date, status FROM tasks")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tasks"})
        return
    }
    defer rows.Close()

    for rows.Next() {
        var task Task
        if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tasks"})
            return
        }
        tasks = append(tasks, task)
    }

    c.JSON(http.StatusOK, tasks)
}

func welcome(c *gin.Context) {
    c.String(http.StatusOK, "Welcome to the Task Manager API!")
}
