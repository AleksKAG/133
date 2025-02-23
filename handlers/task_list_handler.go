package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"go_final_project/models"
)

// Константа для лимита задач
const DefaultTaskLimit = 50

// TaskListResponse структура ответа со списком задач
type TaskListResponse struct {
	Tasks []models.Task `json:"tasks"`
}

// HandleTaskList обрабатывает GET-запросы для получения списка задач
func (h *Handler) HandleTaskList(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем заголовок JSON
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Лимит задач (по умолчанию 50)
	limit := DefaultTaskLimit
	queryLimit := r.URL.Query().Get("limit")
	if queryLimit != "" {
		if parsedLimit, err := strconv.Atoi(queryLimit); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		} else {
			log.Printf("[ОШИБКА] Неверный параметр 'limit': %s", queryLimit)
			writeError(w, "Неверный параметр 'limit'")
			return
		}
	}

	// Выполняем запрос к базе данных
	rows, err := h.DB.Query(
		"SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?",
		limit,
	)
	if err != nil {
		log.Printf("[ОШИБКА] Не удалось выполнить запрос к базе данных: %v", err)
		writeError(w, "Failed to retrieve tasks")
		return
	}
	defer rows.Close()

	// Читаем данные из результата запроса
	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		var id int64 // SQLite возвращает id в виде INTEGER
		err := rows.Scan(&id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			log.Printf("[ОШИБКА] Не удалось разобрать задачу: %v", err)
			writeError(w, "Failed to parse tasks")
			return
		}
		// Преобразуем id в строку
		task.ID = strconv.FormatInt(id, 10)
		tasks = append(tasks, task)
	}

	// Обрабатываем ошибку после завершения итерации
	if err := rows.Err(); err != nil {
		log.Printf("[ОШИБКА] Ошибка при итерации по строкам: %v", err)
		writeError(w, "Error iterating over rows")
		return
	}

	// Если задач нет, возвращаем пустой список
	if tasks == nil {
		tasks = []models.Task{}
	}

	// Формируем и отправляем JSON-ответ
	response := TaskListResponse{Tasks: tasks}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ОШИБКА] Не удалось закодировать задачи в JSON: %v", err)
		writeError(w, "Failed to encode tasks")
	}
}
