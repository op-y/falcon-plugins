package main

import (
	"database/sql"
	"errors"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type Instance struct {
	Host string
	Port string
	Tags string
	Role string
}

var nilDbErr = errors.New("db is nil")

func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, nilDbErr
	}
	return db, nil
}

func GetInstances(db *sql.DB) ([]*Instance, error) {
	instances := []*Instance{}

	sql := "SELECT db_servers_redis.host AS host, db_servers_redis.port AS port, db_servers_redis.tags AS tags, redis_status.redis_role AS role FROM db_servers_redis JOIN redis_status ON db_servers_redis.host=redis_status.host AND db_servers_redis.port=redis_status.port WHERE db_servers_redis.monitor=1"
	stmtSel, err := db.Prepare(sql)
	if err != nil {
		log.Printf("failed to prepare SQL: %s", err.Error())
		return nil, err
	}
	defer stmtSel.Close()

	rows, err := stmtSel.Query()
	if err != nil {
		log.Printf("failed to execute SQL: %s", err.Error())
		return nil, err
	}

	for rows.Next() {
		var host string
		var port string
		var tags string
		var role string

		if err := rows.Scan(&host, &port, &tags, &role); err != nil {
			log.Printf("failed to scan a row: %s", err.Error())
			return nil, err
		}

		instance := &Instance{
			Host: host,
			Port: port,
			Tags: tags,
			Role: role}
		instances = append(instances, instance)
	}
	stmtSel.Close()

	return instances, nil
}

func UpdateStatus(db *sql.DB, tag, host, port, role string, alive, memory_used_percent, max_clients, connected_clients, blocked_clients int64) error {
	is_exist := false

	sql_select := "SELECT * FROM tbl_codis_monitor WHERE redis_ip=? AND port=?"
	stmt_select, err := db.Prepare(sql_select)
	if err != nil {
		log.Printf("failed to prepare SQL: %s", err.Error())
		return err
	}
	defer stmt_select.Close()

	rows, err := stmt_select.Query(host, port)
	if err != nil {
		log.Printf("failed to execute SQL: %s", err.Error())
		return err
	}

	for rows.Next() {
		is_exist = true
	}

	if is_exist {
		sql_update := "UPDATE tbl_codis_monitor SET online_status=?,mem_used_pct=?,max_conns=?,clients=?,blocked_clients=?  WHERE redis_ip=? AND port=?"
		stmt_update, err := db.Prepare(sql_update)
		if err != nil {
			log.Printf("failed to prepare SQL: %s", err.Error())
			return err
		}
		defer stmt_update.Close()

		result, err := stmt_update.Exec(alive, memory_used_percent, max_clients, connected_clients, blocked_clients, host, port)
		if err != nil {
			log.Printf("failed to update monitor data: %s", err.Error())
			return err
		}
		affect, err := result.RowsAffected()
		if err != nil {
			log.Printf("failed to get affected count: %s", err.Error())
			return err
		}
		log.Printf("update %d row(s)", affect)
	} else {
		sql_insert := "INSERT INTO tbl_codis_monitor(tag,redis_ip,port,role,online_status,mem_used_pct,clients,blocked_clients,max_conns) VALUES(?,?,?,?,?,?,?,?,?)"
		stmt_insert, err := db.Prepare(sql_insert)
		if err != nil {
			log.Printf("failed to prepare SQL: %s", err.Error())
			return err
		}
		defer stmt_insert.Close()

		result, err := stmt_insert.Exec(tag, host, port, role, alive, memory_used_percent, connected_clients, blocked_clients, max_clients)
		if err != nil {
			log.Printf("failed to insert monitor data: %s", err.Error())
			return err
		}
		last_id, err := result.LastInsertId()
		if err != nil {
			log.Printf("failed to get last id: %s", err.Error())
			return err
		}
		log.Printf("insert [%d] row", last_id)
	}
	return nil
}
