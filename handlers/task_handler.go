package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"go_final_project/constants"
	"go_final_project/db"
	"go_final_project/models"
	"go_final_project/utils"
)

// HandleTask обрабатывает запросы API для задач
func (h *Handler) HandleTask(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Обработка запроса: %s %s", r.Method, r.URL.Path)
	switch r.Method {
	case http.MethodPost:
		h.addTask(w, r)
	case http.MethodGet:
		h.getTask(w, r)
	case http.MethodPut:
		h.editTask(w, r)
	case http.MethodDelete:
		h.deleteTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Printf("[WARN] Метод %s не поддерживается", r.Method)
	}
}

// addTask добавляет задачу
func (h *Handler) addTask(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] Добавление новой задачи")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var task models.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		log.Printf("[ERROR] Неверный формат JSON, ошибка: %v", err)
		writeError(w, "Неверный формат JSON")
		return
	}

	now := utils.NormalizeDate(time.Now())

	if task.Date == "" {
		task.Date = now.Format(constants.DateFormat)
	} else {
		parsedDate, err := time.Parse(constants.DateFormat, task.Date)
		if err != nil {
			log.Printf("[ERROR] Неверный формат даты: %s, ошибка: %v", task.Date, err)
			writeError(w, "Неверный формат даты (ожидается YYYYMMDD)")
			return
		}

		if parsedDate.Before(now) || parsedDate.Equal(now) {
			if task.Repeat == "" {
				task.Date = now.Format(constants.DateFormat)
			} else {
				task.Date, err = utils.NextDate(now, task.Date, task.Repeat)
				if err != nil {
					log.Printf("[ERROR] Некорректное правило повторения: %s, ошибка: %v", task.Repeat, err)
					writeError(w, "Некорректное правило повторения")
					return
				}
			}
		}
	}

	if task.Title == "" {
		log.Println("[ERROR] Не указан заголовок задачи")
		writeError(w, "Не указан заголовок задачи")
		return
	}

	id, err := db.AddTask(h.DB, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		log.Printf("[ERROR] Ошибка при добавлении задачи, заголовок: %s, ошибка: %v", task.Title, err)
		writeError(w, "Не удалось добавить задачу")
		return
	}
	log.Printf("[INFO] Задача добавлена с ID %d", id)
	response := map[string]any{"id": strconv.FormatInt(id, 10)}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ERROR] Ошибка при формировании ответа, ID: %d, ошибка: %v", id, err)
		writeError(w, "Ошибка при формировании ответа")
	}
}

// getTask возвращает данные задачи по идентификатору
func (h *Handler) getTask(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] Получение задачи")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	if id == "" {
		log.Println("[ERROR] Не указан идентификатор задачи")
		writeError(w, "Не указан идентификатор задачи")
		return
	}

	taskID, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("[ERROR] Неверный формат идентификатора задачи: %s", id)
		writeError(w, "Идентификатор задачи должен быть числом")
		return
	}

	task, err := db.GetTaskByID(h.DB, taskID)
	if err != nil {
		log.Printf("[ERROR] Ошибка при получении задачи, ID: %d, ошибка: %v", taskID, err)
		writeError(w, "Ошибка при получении задачи")
		return
	}

	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("[ERROR] Ошибка при формировании ответа, ID: %d, ошибка: %v", taskID, err)
		writeError(w, "Ошибка при формировании ответа")
	}
}

// editTask обновляет параметры задачи
func (h *Handler) editTask(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] Обновление задачи")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		log.Printf("[ERROR] Неверный формат JSON, ошибка: %v", err)
		writeError(w, "Неверный формат JSON")
		return
	}

	if task.ID == "" {
		log.Println("[ERROR] Не указан идентификатор задачи")
		writeError(w, "Не указан идентификатор задачи")
		return
	}

	if task.Date != "" {
		if _, err := time.Parse(constants.DateFormat, task.Date); err != nil {
			log.Printf("[ERROR] Неверный формат даты: %s, ошибка: %v", task.Date, err)
			writeError(w, "Неверный формат даты (ожидается YYYYMMDD)")
			return
		}
	} else {
		task.Date = utils.NormalizeDate(time.Now()).Format(constants.DateFormat)
	}

	if task.Title == "" {
		log.Println("[ERROR] Заголовок задачи обязателен")
		writeError(w, "Заголовок задачи обязателен")
		return
	}

	rowsAffected, err := db.UpdateTask(h.DB, task)
	if err != nil || rowsAffected == 0 {
		log.Printf("[ERROR] Ошибка при обновлении задачи, ID: %s, ошибка: %v", task.ID, err)
		writeError(w, "Задача не найдена или не удалось обновить")
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]any{}); err != nil {
		log.Printf("[ERROR] Ошибка при отправке ответа, ID: %s, ошибка: %v", task.ID, err)
		writeError(w, "Ошибка при отправке ответа")
	}
}

// HandleTaskDone завершает задачу
func (h *Handler) HandleTaskDone(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] Завершение задачи")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	if id == "" {
		log.Println("[ERROR] Не указан идентификатор задачи")
		writeError(w, "Не указан идентификатор задачи")
		return
	}

	taskID, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("[ERROR] Неверный формат идентификатора задачи: %s", id)
		writeError(w, "Идентификатор задачи должен быть числом")
		return
	}

	// Получаем задачу из базы данных
	task, err := db.GetTaskByID(h.DB, taskID)
	if err != nil {
		log.Printf("[ERROR] Ошибка при получении задачи, ID: %d, ошибка: %v", taskID, err)
		writeError(w, "Ошибка при получении задачи")
		return
	}

	if task.Repeat == "" {
		// Если задача одноразовая, удаляем её
		_, err = db.DeleteTask(h.DB, taskID)
		if err != nil {
			log.Printf("[ERROR] Не удалось удалить задачу, ID: %d, ошибка: %v", taskID, err)
			writeError(w, "Не удалось удалить задачу")
			return
		}
	} else {
		// Если задача повторяющаяся, обновляем дату
		now := utils.NormalizeDate(time.Now())
		nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			log.Printf("[ERROR] Ошибка при расчёте следующей даты, ID: %d, ошибка: %v", taskID, err)
			writeError(w, "Ошибка при расчёте следующей даты")
			return
		}

		task.Date = nextDate
		_, err = db.UpdateTask(h.DB, *task)
		if err != nil {
			log.Printf("[ERROR] Не удалось обновить задачу, ID: %d, ошибка: %v", taskID, err)
			writeError(w, "Не удалось обновить задачу")
			return
		}
	}

	if err := json.NewEncoder(w).Encode(map[string]any{}); err != nil {
		log.Printf("[ERROR] Ошибка при отправке ответа, ID: %d, ошибка: %v", taskID, err)
		writeError(w, "Ошибка при отправке ответа")
	}
}

// deleteTask удаляет задачу по идентификатору
func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] Удаление задачи")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	if id == "" {
		log.Println("[ERROR] Не указан идентификатор задачи")
		writeError(w, "Не указан идентификатор задачи")
		return
	}

	taskID, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("[ERROR] Неверный формат идентификатора задачи: %s", id)
		writeError(w, "Идентификатор задачи должен быть числом")
		return
	}

	// Удаляем задачу из базы данных через db.DeleteTask
	rowsAffected, err := db.DeleteTask(h.DB, taskID)
	if err != nil {
		log.Printf("[ERROR] Ошибка при удалении задачи, ID: %d, ошибка: %v", taskID, err)
		writeError(w, "Не удалось удалить задачу")
		return
	}

	// Проверяем, была ли удалена задача
	if rowsAffected == 0 {
		log.Printf("[WARNING] Попытка удалить несуществующую задачу, ID: %d", taskID)
		writeError(w, "Задача не найдена")
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]any{}); err != nil {
		log.Printf("[ERROR] Ошибка при отправке ответа, ID: %d, ошибка: %v", taskID, err)
		writeError(w, "Ошибка при отправке ответа")
	}
}

// writeError отправляет сообщение об ошибке в формате JSON
func writeError(w http.ResponseWriter, message string) {
	log.Printf("[ERROR] %s", message)
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{"error": message})
}
