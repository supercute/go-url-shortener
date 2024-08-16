package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/dgraph-io/badger/v3"
)

type Storage struct {
	db         *badger.DB
	adminToken string
}

// LinkData содержит информацию о сокращенной ссылке и её владельце.
type LinkData struct {
	URL   string `json:"url"`
	Owner string `json:"owner"`
}

// NewStorage инициализирует новую базу данных Badger и загружает админский токен из конфигурационного файла.
func NewStorage() (*Storage, error) {
	opts := badger.DefaultOptions("./badger").WithInMemory(false)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	storage := &Storage{db: db}
	if err := storage.loadAdminToken(); err != nil {
		return nil, err
	}

	return storage, nil
}

// Close закрывает соединение с базой данных.
func (s *Storage) Close() {
	if err := s.db.Close(); err != nil {
		log.Println("Error closing the database:", err)
	}
}

// loadAdminToken загружает админский токен из конфигурационного файла.
func (s *Storage) loadAdminToken() error {
	file, err := os.Open("./configs/config.json")
	if err != nil {
		return err
	}
	defer file.Close()

	var config struct {
		AdminToken string `json:"admin_token"`
	}

	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return err
	}

	s.adminToken = config.AdminToken
	return nil
}

// GetAdminToken возвращает админский токен.
func (s *Storage) GetAdminToken() string {
	return s.adminToken
}

// SaveLink сохраняет сокращенную ссылку и соответствующую ей исходную ссылку, а также владельца ссылки.
func (s *Storage) SaveLink(name, link, owner string) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte("link:" + name))
		if err == nil {
			return fmt.Errorf("link name %s already exists", name)
		} else if err != badger.ErrKeyNotFound {
			return err
		}

		linkData := LinkData{
			URL:   link,
			Owner: owner,
		}

		data, err := json.Marshal(linkData)
		if err != nil {
			return err
		}

		return txn.Set([]byte("link:"+name), data)
	})
	return err
}

// GetLink возвращает исходную ссылку по её сокращенному имени.
func (s *Storage) GetLink(name string) (string, error) {
	var linkData LinkData
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("link:" + name))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &linkData)
		})
	})
	if err != nil {
		return "", err
	}
	return linkData.URL, nil
}

// NameExists проверяет, существует ли уже имя для сокращенной ссылки.
func (s *Storage) NameExists(name string) bool {
	err := s.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte("link:" + name))
		return err
	})
	return err == nil
}

// EmailExists проверяет, существует ли уже email в базе данных.
func (s *Storage) EmailExists(email string) bool {
	err := s.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte("user:" + email))
		return err
	})
	return err == nil
}

// SaveToken сохраняет токен, связанный с email.
func (s *Storage) SaveToken(email, token string) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("user:"+email), []byte(token))
	})
	return err
}

// DeleteUser удаляет пользователя и все его ссылки из базы данных.
func (s *Storage) DeleteUser(email string) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		// Удаляем пользователя
		if err := txn.Delete([]byte("user:" + email)); err != nil {
			return err
		}

		// Удаляем все ссылки пользователя
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek([]byte("link:")); it.ValidForPrefix([]byte("link:")); it.Next() {
			item := it.Item()
			var linkData LinkData
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &linkData)
			})
			if err != nil {
				return err
			}

			if linkData.Owner == email {
				if err := txn.Delete(item.Key()); err != nil {
					return err
				}
			}
		}
		return nil
	})
	return err
}

// GetAllUsers возвращает список всех пользователей с количеством сгенерированных ссылок для каждого.
func (s *Storage) GetAllUsers() map[string]int {
	users := make(map[string]int)
	s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		// Получаем всех пользователей
		for it.Seek([]byte("user:")); it.ValidForPrefix([]byte("user:")); it.Next() {
			item := it.Item()
			key := string(item.Key())
			email := key[5:]
			users[email] = 0
		}

		// Считаем количество ссылок для каждого пользователя
		for it.Seek([]byte("link:")); it.ValidForPrefix([]byte("link:")); it.Next() {
			item := it.Item()
			var linkData LinkData
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &linkData)
			})
			if err != nil {
				continue
			}
			if count, exists := users[linkData.Owner]; exists {
				users[linkData.Owner] = count + 1
			}
		}

		return nil
	})
	return users
}
