package database

import (
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

type Event struct {
	ID        int
	Type      string // "log", "status"
	Message   string // Contenido del log o estado ("online", "offline")
	Timestamp time.Time
}

type DB struct {
	conn *sql.DB
}

func Init() *DB {
	conn, err := sql.Open("sqlite", "houseauto.db")
	if err != nil {
		log.Fatalf("[-] Error al abrir base de datos: %v", err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		message TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := conn.Exec(createTableQuery); err != nil {
		log.Fatalf("[-] Error al crear tabla de eventos: %v", err)
	}

	log.Println("[+] Base de datos SQLite inicializada correctamente")
	return &DB{conn: conn}
}

func (db *DB) SaveEvent(eventType, message string) error {
	_, err := db.conn.Exec("INSERT INTO events (type, message) VALUES (?, ?)", eventType, message)
	if err != nil {
		log.Printf("[-] Error al guardar evento en DB: %v", err)
	}
	return err
}

func (db *DB) GetRecentEvents(limit int) ([]Event, error) {
	rows, err := db.conn.Query("SELECT id, type, message, timestamp FROM events ORDER BY timestamp DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Type, &e.Message, &e.Timestamp); err != nil {
			log.Printf("[-] Error leyendo evento: %v", err)
			continue
		}
		events = append(events, e)
	}

	// Como los pedimos DESC, los más nuevos están primero.
	// Si queremos mandarlos al frontend en orden cronológico (los más viejos arriba), revertimos:
	for i, j := 0, len(events)-1; i < j; i, j = i+1, j-1 {
		events[i], events[j] = events[j], events[i]
	}

	return events, nil
}
