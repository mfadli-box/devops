package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-routeros/routeros/v3"
	"github.com/joho/godotenv"
)

var PgSQL *sql.DB

type MikrotikRouter struct {
	ID        string
	Name      string
	IPAddress string
	Port      int
	Username  string
	Password  string
}

func StartNMSPolling(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			routers, err := fetchActiveRouters()
			if err != nil {
				log.Printf("Gagal mengambil data router dari DB : %v", err)
				continue
			}

			for _, r := range routers {
				go pollRouterMetrics(r)
			}
		case <-ctx.Done():
			return
		}
	}
}

func fetchActiveRouters() ([]MikrotikRouter, error) {
	query := `
		SELECT	id, name, ip_address, api_port, username, password
		FROM	"ict_mikrotik_router"`
	rows, err := PgSQL.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routers []MikrotikRouter
	for rows.Next() {
		var r MikrotikRouter
		if err := rows.Scan(&r.ID, &r.Name, &r.IPAddress, &r.Port, &r.Username, &r.Password); err != nil {
			return nil, err
		}
		routers = append(routers, r)
	}
	return routers, nil
}

func pollRouterMetrics(r MikrotikRouter) {
	addr := fmt.Sprintf("%s:%d", r.IPAddress, r.Port)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := routeros.DialContext(ctx, addr, r.Username, r.Password)

	if err != nil {
		log.Printf("Router %s (%s) OFFLINE/UNREACHABLE: %v", r.Name, r.IPAddress, err)
		updateRouterStatus(r.ID, "UNREACHABLE")
		return
	}
	defer client.Close()

	updateRouterStatus(r.ID, "ONLINE")

	res, err := client.Run("/system/resource/print")
	if err == nil && len(res.Re) > 0 {
		cpu, _ := strconv.Atoi(res.Re[0].Map["cpu-load"])
		freeMem, _ := strconv.ParseInt(res.Re[0].Map["free-memory"], 10, 64)
		uptime := res.Re[0].Map["uptime"]
		query_a := `
			INSERT INTO "ict_mikrotik_metric" (
				id, router_id, cpu_load, free_memory, uptime, timestamp
			) VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())`
		_, _ = PgSQL.Exec(query_a, r.ID, cpu, freeMem, uptime)
	}

	interfaces, err := client.Run("/interface/print")
	if err == nil {
		for _, iface := range interfaces.Re {
			ifaceName := iface.Map["name"]
			ifaceType := iface.Map["type"]
			macAddr := iface.Map["mac-address"]
			comment := iface.Map["comment"]

			var ifaceID string
			query_b := `
				INSERT INTO "ict_mikrotik_interface" (
					id, router_id, name, type, mac_address, description, updated_at
				) VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW())
				ON CONFLICT (router_id, name) DO UPDATE 
				SET type=$3, mac_address=$4, description=$5, updated_at=NOW()
				RETURNING id`
			err := PgSQL.QueryRow(query_b,
				r.ID, ifaceName, ifaceType, macAddr, comment,
			).Scan(&ifaceID)

			if err == nil {
				txBytes, _ := strconv.ParseInt(iface.Map["tx-byte"], 10, 64)
				rxBytes, _ := strconv.ParseInt(iface.Map["rx-byte"], 10, 64)
				txPackets, _ := strconv.ParseInt(iface.Map["tx-packet"], 10, 64)
				rxPackets, _ := strconv.ParseInt(iface.Map["rx-packet"], 10, 64)
				query_c := `
					INSERT INTO "ict_mikrotik_intmet" (
						id, interface_id, tx_bytes, rx_bytes, tx_packets, rx_packets, timestamp
					) VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW())`
				_, _ = PgSQL.Exec(query_c,
					ifaceID, txBytes, rxBytes, txPackets, rxPackets,
				)
			}
		}
	}
}

func updateRouterStatus(routerID string, status string) {
	query := `
		UPDATE	"ict_mikrotik_router"
		SET		status = $1, last_seen = NOW(), updated_at = NOW()
		WHERE	id = $2`
	_, _ = PgSQL.Exec(query, status, routerID)
}

func PushAddressList(routerID string, listName string, address string, comment string) error {
	var r MikrotikRouter
	query_a := `
		SELECT	id, ip_address, api_port, username, password
		FROM	"ict_mikrotik_router"
		WHERE	id = $1`
	err := PgSQL.QueryRow(query_a, routerID).Scan(
		&r.ID,
		&r.IPAddress,
		&r.Port,
		&r.Username,
		&r.Password)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", r.IPAddress, r.Port)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := routeros.DialContext(ctx, addr, r.Username, r.Password)
	if err != nil {
		return err
	}
	defer client.Close()

	cmd := []string{"/ip/firewall/address-list/add", "=list=" + listName, "=address=" + address}
	if comment != "" {
		cmd = append(cmd, "=comment="+comment)
	}

	res, err := client.RunArgs(cmd)
	if err != nil {
		return err
	}

	query_b := `
		INSERT INTO "ict_mikrotik_address_list" (
			id, router_id, mikrotik_id, list_name, address, comment, created_at, updated_at
		) VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW(), NOW())`
	mikrotikID := res.Done.Map["ret"]
	_, err = PgSQL.Exec(query_b, routerID, mikrotikID, listName, address, comment)
	return err
}

func PushFilterRule(routerID string, chain string, action string, protocol string, dstPort string, comment string, position int) error {
	var r MikrotikRouter
	query_a := `
		SELECT	id, ip_address, api_port, username, password
		FROM	"ict_mikrotik_router"
		WHERE	id = $1`
	err := PgSQL.QueryRow(query_a, routerID).Scan(
		&r.ID,
		&r.IPAddress,
		&r.Port,
		&r.Username,
		&r.Password)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", r.IPAddress, r.Port)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := routeros.DialContext(ctx, addr, r.Username, r.Password)
	if err != nil {
		return err
	}
	defer client.Close()

	cmd := []string{"/ip/firewall/filter/add", "=chain=" + chain, "=action=" + action, fmt.Sprintf("=place-before=%d", position)}
	if protocol != "" {
		cmd = append(cmd, "=protocol="+protocol)
	}
	if dstPort != "" {
		cmd = append(cmd, "=dst-port="+dstPort)
	}
	if comment != "" {
		cmd = append(cmd, "=comment="+comment)
	}

	res, err := client.RunArgs(cmd)
	if err != nil {
		return err
	}

	mikrotikID := res.Done.Map["ret"]
	query_b := `
		INSERT INTO "ict_mikrotik_filter_rule" (
			id, router_id, mikrotik_id, chain, action, protocol, dst_port, comment, position, created_at, updated_at
		) VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())`
	_, err = PgSQL.Exec(query_b, routerID, mikrotikID, chain, action, protocol, dstPort, comment, position)
	return err
}

func DeleteAddressList(id string) error {
	var routerIP, username, password, mikrotikID string
	var port int

	query_a := `
		SELECT	r.ip_address, r.api_port, r.username, r.password, al.mikrotik_id 
		FROM	"ict_mikrotik_address_list" al
		JOIN	"ict_mikrotik_router" r ON al.router_id = r.id
		WHERE	al.id = $1`
	err := PgSQL.QueryRow(query_a, id).Scan(
		&routerIP,
		&port,
		&username,
		&password,
		&mikrotikID)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := routeros.DialContext(ctx, fmt.Sprintf("%s:%d", routerIP, port), username, password)
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.Run("/ip/firewall/address-list/remove", "=.id="+mikrotikID)
	if err != nil {
		return err
	}

	query_b := `
		DELETE	FROM "ict_mikrotik_address_list"
		WHERE	id = $1`
	_, err = PgSQL.Exec(query_b, id)
	return err
}

func main() {
	godotenv.Load()
	PG_Host := os.Getenv("PG_HOST")
	PG_Port := os.Getenv("PG_PORT")
	PG_User := os.Getenv("PG_USER")
	PG_Pass := os.Getenv("PG_PASS")
	PG_Data := os.Getenv("PG_DATA")
	IS_Pool := os.Getenv("IS_POOL")
	var dsn string
	dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		PG_Host, PG_Port, PG_User, PG_Pass, PG_Data)

	var err error
	PgSQL, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Gagal membuka basis data %v", err)
	}

	if err = PgSQL.Ping(); err != nil {
		log.Fatalf("Basis data tidak merespon %v", err)
	}

	if IS_Pool == "true" {
		PgSQL.SetMaxOpenConns(100)
		PgSQL.SetMaxIdleConns(10)
	} else {
		PgSQL.SetMaxOpenConns(50)
		PgSQL.SetMaxIdleConns(25)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go StartNMSPolling(ctx, 30*time.Second)
}
