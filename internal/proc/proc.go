package proc

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const CONNECTION_STATE_LISTEN = "0A"
const CONNECTION_STATE_CLOSE = "07"
const PROC_ROOT = "/proc/"

type Entity struct {
	inode uint64
	pid   int
	fd    int
	name  string
}

func GetUserEntities() (map[uint64]*Entity, error) {
	user_entities := make(map[uint64]*Entity)

	proc_dir, err := os.ReadDir(PROC_ROOT)
	if err != nil {
		return user_entities, err
	}
	for _, proc_file := range proc_dir {
		var process_name string

		proc_file_name := proc_file.Name()
		proc_file_name_int, err := strconv.Atoi(proc_file_name)
		if err != nil {
			slog.Debug("Error converting to integer: %v", "err", err)
			continue
		}

		proc_fd_path := filepath.Join(PROC_ROOT, proc_file_name, "fd")
		proc_fds, err := os.ReadDir(proc_fd_path)
		if err != nil {
			return user_entities, err
		}

		for _, fd := range proc_fds {
			fd_name := fd.Name()
			fd_name_int, err := strconv.Atoi(fd_name)
			if err != nil {
				return user_entities, err
			}

			fd_path := filepath.Join(proc_fd_path, fd_name)
			link, err := os.Readlink(fd_path)
			if err != nil {
				slog.Debug("Error read link: %v", "err", err)
				continue
			}

			if strings.Contains(link, "socket:[") {
				var inode uint64
				_, err := fmt.Sscanf(link, "socket:[%d]", &inode)
				if err != nil {
					return user_entities, err
				}

				if process_name == "" {
					first_char := byte('(')
					last_char := byte(')')
					proc_pid_stat := filepath.Join(PROC_ROOT, proc_file_name, "stat")
					if file, err := os.Open(proc_pid_stat); err != nil {
						return user_entities, err
					} else {
						if data, err := io.ReadAll(file); err != nil {
							return user_entities, err
						} else {
							first_index := bytes.IndexByte(data, first_char)
							last_index := first_index + bytes.IndexByte(data[first_index:], last_char)
							process_name = string(data[first_index+1 : last_index])
						}
					}
				}

				user_entities[inode] = &Entity{
					inode,
					proc_file_name_int,
					fd_name_int,
					process_name,
				}
			}
		}

	}
	return user_entities, nil
}

func ShowNetTcp(user_entities map[uint64]*Entity) error {

	proc_net_tcp_path := filepath.Join(PROC_ROOT, "net", "tcp")
	if file, err := os.Open(proc_net_tcp_path); err != nil {
		return err
	} else {
		if bytes, err := io.ReadAll(file); err != nil {
			return err
		} else {
			tcp_data := strings.Split(strings.TrimSpace(string(bytes)), "\n")[1:]
			for _, socket := range tcp_data {
				socket_data := strings.Fields(socket)
				if socket_data[3] == CONNECTION_STATE_LISTEN {
					var l1, l2, l3, l4 uint8
					var lp1 uint16

					_, err := fmt.Sscanf(socket_data[1], "%2x%2x%2x%2x:%x", &l1, &l2, &l3, &l4, &lp1)
					if err != nil {
						slog.Error("Error scanning address:port", "err", err, "file", proc_net_tcp_path)
					}

					remote_address_port := "0.0.0.0:*"

					var transmit_queue, receive_queue uint
					_, err = fmt.Sscanf(socket_data[4], "%8x:%8x", &transmit_queue, &receive_queue)
					if err != nil {
						slog.Error("Error scanning tq:rq", "err", err, "file", proc_net_tcp_path)
					}

					inode, err := strconv.ParseUint(socket_data[9], 10, 64)
					if err != nil {
						slog.Error("Error scanning inode", "err", err, "file", proc_net_tcp_path)
					}
					user_entity := user_entities[inode]
					// TODO: Поправить форматирование
					fmt.Printf(
						"%-10s %-6d %-6d %4d.%d.%d.%d:%-5d %4s (%q,pid=%d,fd=%d)\n",
						"LISTEN",
						transmit_queue,
						receive_queue,
						l4, l3, l2, l1, lp1,
						remote_address_port,
						user_entity.name,
						user_entity.pid,
						user_entity.fd,
					)
				}
			}

		}
	}

	return nil
}

func ShowNetUdp(user_entities map[uint64]*Entity) error {
	proc_net_udp_path := filepath.Join(PROC_ROOT, "net", "udp")
	if file, err := os.Open(proc_net_udp_path); err != nil {
		return err
	} else {
		if bytes, err := io.ReadAll(file); err != nil {
			return err
		} else {
			udp_data := strings.Split(strings.TrimSpace(string(bytes)), "\n")[1:]
			for _, socket := range udp_data {
				socket_data := strings.Fields(socket)
				if socket_data[3] == CONNECTION_STATE_CLOSE {
					var l1, l2, l3, l4 uint8
					var lp1 uint16

					_, err := fmt.Sscanf(socket_data[1], "%2x%2x%2x%2x:%x", &l1, &l2, &l3, &l4, &lp1)
					if err != nil {
						slog.Error("Error scanning address:port", "err", err, "file", proc_net_udp_path)
					}

					remote_address_port := "0.0.0.0:*"

					var transmit_queue, receive_queue uint
					_, err = fmt.Sscanf(socket_data[4], "%8x:%8x", &transmit_queue, &receive_queue)
					if err != nil {
						slog.Error("Error scanning tq:rq", "err", err, "file", proc_net_udp_path)
					}

					inode, err := strconv.ParseUint(socket_data[9], 10, 64)
					if err != nil {
						slog.Error("Error scanning inode", "err", err, "file", proc_net_udp_path)
					}

					user_entity, exists := user_entities[inode]
					process := ""
					if exists {
						process = fmt.Sprintf("(%q,pid=%d,fd=%d)", user_entity.name, user_entity.pid, user_entity.fd)
					}

					// TODO: Поправить форматирование
					fmt.Printf(
						"%-10s %-6d %-6d %4d.%d.%d.%d:%-5d %-15s %-10s\n",
						"UNCONN",
						transmit_queue,
						receive_queue,
						l4, l3, l2, l1, lp1,
						remote_address_port,
						process,
					)
				}
			}

		}
	}
	return nil
}
