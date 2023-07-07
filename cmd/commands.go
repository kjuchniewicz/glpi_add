package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kjuchniewicz/glpi_add/pkg"
	"github.com/kjuchniewicz/glpi_add/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewRootCommand() *cobra.Command {
	desc := ""
	who := 0
	actiontime := 1
	version := false
	showConfig := false
	setConfig := false

	glpiCmd := &cobra.Command{
		Use:   "glpi",
		Short: "Szybciutko a działa",
		Long:  "\nUmożliwia dodanie powiązanego zgłoszenia, zadania, rozwiązania i jego akceptacji.",

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return utils.InitializeConfig(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			if version {
				printVersion()
			} else if showConfig {
				showCurrentConfig(cmd)
			} else if setConfig {
				setNewConfig()
			} else {
				dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", "glpi", "Barakuda7200", "192.168.1.110", 3306, "glpi944")
				db, err := sql.Open("mysql", dsn)
				if err != nil {
					log.Fatalf("nie można utworzyć połączenia: %s", err)
				}
				defer db.Close()

				db.SetConnMaxLifetime(time.Minute * 3)
				db.SetMaxOpenConns(10)
				db.SetMaxIdleConns(10)

				mail := ""
				if who == 43 {
					mail = "k.juchniewicz@zarna.pl"
				} else if who == 53 {
					mail = "m.bilkiewicz@zarna.pl"
				} else if who == 298 {
					mail = "m.jurgielewicz@zarna.pl"
				}

				query_ticket := `insert into
					glpi_tickets
					(
						name,                         -- Zadanie
						date,                         -- CURRENT_TIMESTAMP
						date_creation,                -- CURRENT_TIMESTAMP
						closedate,                    -- CURRENT_TIMESTAMP
						solvedate,                    -- CURRENT_TIMESTAMP
						takeintoaccountdate,          -- CURRENT_TIMESTAMP
						date_mod,                     -- CURRENT_TIMESTAMP
						status,                       -- 6
						users_id_lastupdater,         -- KJ 43 MB 53 MJ 298
						users_id_recipient,           -- KJ 43 MB 53 MJ 298
						requesttypes_id,              -- 1
						content,                      -- Zadanie
						urgency, -- 3
						impact, -- 3
						priority, -- 3
						slas_id_ttr, -- 4
						slas_id_tto, -- 3
						slalevels_id_ttr, -- 1
						time_to_resolve,              -- TIMESTAMPADD(DAY,9,CURRENT_TIMESTAMP)
						time_to_own,                  -- TIMESTAMPADD(DAY,1,CURRENT_TIMESTAMP)
						begin_waiting_date, -- CURRENT_TIMESTAMP
						close_delay_stat, -- 120
						solve_delay_stat, -- 60
						takeintoaccount_delay_stat,   -- 50
						actiontime                   -- czas w [godziny]
					)
				values
				(
					?,
					CURRENT_TIMESTAMP,
					CURRENT_TIMESTAMP,
					TIMESTAMPADD(MINUTE,120,CURRENT_TIMESTAMP),
					TIMESTAMPADD(MINUTE,60,CURRENT_TIMESTAMP),
					TIMESTAMPADD(MINUTE,60,CURRENT_TIMESTAMP),
					TIMESTAMPADD(MINUTE,50,CURRENT_TIMESTAMP),
					6,
					?,
					?,
					1,
					?,
					3,
					3,
					3,
					4,
					3,
					1,
					TIMESTAMPADD(DAY,9,CURRENT_TIMESTAMP),
					TIMESTAMPADD(DAY,1,CURRENT_TIMESTAMP),
					CURRENT_TIMESTAMP,
					120,
					60,
					50,
					?)`
				insertResult, err := db.ExecContext(context.Background(), query_ticket, desc, who, who, desc, actiontime*3600)
				if err != nil {
					log.Fatalf("nie można dodać zgłoszenia: %s", err)
				}
				ticket_id, err := insertResult.LastInsertId()
				if err != nil {
					log.Fatalf("niemożna uzyskać numeru zgłoszenia: %s", err)
				}
				fmt.Printf("Ticket: %d", ticket_id)

				query_task := `insert into
						glpi_tickettasks
						(
							uuid,                                       -- UUID()
							tickets_id,                                 -- id z założonego ticket'u
							date,                                       -- TIMESTAMPADD(MINUTE,55,CURRENT_TIMESTAMP)
							users_id,                                   -- KJ 43 MB 53 MJ 298
							users_id_editor,                            -- KJ 43 MB 53 MJ 298
							users_id_tech,                              -- KJ 43 MB 53 MJ 298
							content,                                    -- Zadanie
							actiontime,                                 -- czas w [godziny]
							state,                                      -- 2
							date_mod,                                   -- TIMESTAMPADD(MINUTE,55,CURRENT_TIMESTAMP)
							date_creation,                              -- CURRENT_TIMESTAMP
							timeline_position                           -- 1
						)
					values
					(
						UUID(),
						?,
						TIMESTAMPADD(MINUTE,55,CURRENT_TIMESTAMP),
						?,
						?,
						?,
						?,
						?,
						2,
						TIMESTAMPADD(MINUTE,55,CURRENT_TIMESTAMP),
						CURRENT_TIMESTAMP,
						1
					)`
				insertResult, err = db.ExecContext(context.Background(), query_task, ticket_id, who, who, who, desc, actiontime*3600)
				if err != nil {
					log.Fatalf("nie można dodać zadania: %s", err)
				}
				id, err := insertResult.LastInsertId()
				if err != nil {
					log.Fatalf("niemożna uzyskać numeru zadania: %s", err)
				}
				fmt.Printf("Task: %d", id)

				query_users := `insert into
						glpi_tickets_users
						(
							tickets_id,                                 -- id z założonego ticket'u
							users_id,                                   -- KJ 43 MB 53 MJ 298
							type,                                       -- 1 i 2
							use_notification,                           -- 0
							alternative_email                           -- wiadomo
						)
					values
					(
						?,
						?,
						1,
						0,
						?
					),
					(
						?,
						?,
						2,
						0,
						?
					)`
				insertResult, err = db.ExecContext(context.Background(), query_users, ticket_id, who, mail, ticket_id, who, mail)
				if err != nil {
					log.Fatalf("nie można dodać użytkowników: %s", err)
				}
				id, err = insertResult.LastInsertId()
				if err != nil {
					log.Fatalf("niemożna uzyskać numeru użytkowników: %s", err)
				}
				fmt.Printf("Users: %d", id)

				query_solution := `INSERT INTO
						glpi_itilsolutions
						(
							itemtype,                                         -- Ticket
							items_id,                                         -- id z założonego ticket'u
							content,                                          -- &#60;p&#62;Brak komentarzy&#60;/p&#62;
							date_creation,                                    -- TIMESTAMPADD(MINUTE,60,CURRENT_TIMESTAMP),
							date_mod,                                         -- TIMESTAMPADD(MINUTE,60,CURRENT_TIMESTAMP),
							date_approval,                                    -- TIMESTAMPADD(MINUTE,65,CURRENT_TIMESTAMP),
							users_id,                                         -- KJ 43 MB 53 MJ 298
							users_id_approval,                                -- KJ 43 MB 53 MJ 298
							status                                            -- 3
						)
					VALUES
					(
						'Ticket',
						?,
						'&#60;p&#62;Brak komentarzy&#60;/p&#62;',
						TIMESTAMPADD(MINUTE,60,CURRENT_TIMESTAMP),
						TIMESTAMPADD(MINUTE,60,CURRENT_TIMESTAMP),
						TIMESTAMPADD(MINUTE,65,CURRENT_TIMESTAMP),
						?,
						?,
						3
					)`
				insertResult, err = db.ExecContext(context.Background(), query_solution, ticket_id, who, who)
				if err != nil {
					log.Fatalf("nie można dodać rozwiązania: %s", err)
				}
				id, err = insertResult.LastInsertId()
				if err != nil {
					log.Fatalf("niemożna uzyskać numeru rozwiązania: %s", err)
				}
				fmt.Printf("Rozwiązanie: %d", id)

				query_followup := `INSERT INTO
						glpi_itilfollowups
						(
							itemtype,                                         -- Ticket
							items_id,                                         -- id z założonego ticket'u
							date,                                             -- TIMESTAMPADD(MINUTE,70,CURRENT_TIMESTAMP),
							date_creation,                                    -- TIMESTAMPADD(MINUTE,70,CURRENT_TIMESTAMP),
							date_mod,                                         -- TIMESTAMPADD(MINUTE,75,CURRENT_TIMESTAMP),
							users_id,                                         -- KJ 43 MB 53 MJ 298
							content,                                          -- Zatwierdzone rozwiązanie
							requesttypes_id,                                  -- id z założonego ticket'u
							timeline_position                                 -- 1
						)
					VALUES
					(
						'Ticket',
						?,
						TIMESTAMPADD(MINUTE,70,CURRENT_TIMESTAMP),
						TIMESTAMPADD(MINUTE,70,CURRENT_TIMESTAMP),
						TIMESTAMPADD(MINUTE,75,CURRENT_TIMESTAMP),
						?,
						'Zatwierdzone rozwiązanie',
						?,
						1
					)`
				insertResult, err = db.ExecContext(context.Background(), query_followup, ticket_id, who, ticket_id)
				if err != nil {
					log.Fatalf("nie można dodać zatwierdzenia: %s", err)
				}
				id, err = insertResult.LastInsertId()
				if err != nil {
					log.Fatalf("niemożna uzyskać numeru zatwierdzenia: %s", err)
				}
				fmt.Printf("Zatwierdzenie: %d", id)

				// out := cmd.OutOrStdout()
				// fmt.Fprintln(out, "Jesteś", who)
				// fmt.Fprintln(out, "Zrobiłeś:", desc)
				// fmt.Fprintln(out, "Zrobiłeś:", actiontime)
			}
		},
	}

	glpiCmd.Flags().StringVarP(&desc, "opis", "o", "Prace nad sobą", "treść zadania będzie jednocześnie tytułem i opisem zgłoszenia")
	glpiCmd.Flags().IntVarP(&who, "kto", "k", 43, "osoba na którą zostanie założone zgłoszenie i zadanie - [KJ 43 MB 53 MJ 298]")
	glpiCmd.Flags().IntVarP(&actiontime, "czas", "t", 1, "czas poświęcony na zadanie w godzinach")
	glpiCmd.Flags().BoolVarP(&version, "wersja", "w", false, "pokazuje aktualną wersję CLI")
	glpiCmd.Flags().BoolVarP(&showConfig, "konf", "c", false, "pokazuje aktualną konfiguracje")
	glpiCmd.Flags().BoolVarP(&setConfig, "ustaw konf", "u", false, "ustawia nową konfigurację")

	return glpiCmd
}

func printVersion() {
	fmt.Println("Wersja: ", pkg.Version)
}

func showCurrentConfig(cmd *cobra.Command) {
	desc, err := cmd.Flags().GetString("opis")
	if err != nil {
		panic(err)
	}

	who := 43
	who, err = cmd.Flags().GetInt("kto")
	if err != nil {
		panic(err)
	}

	actiontime := 1
	actiontime, err = cmd.Flags().GetInt("czas")
	if err != nil {
		panic(err)
	}

	fmt.Printf("kto: %d\n"+
		"opis: %q\n"+
		"czas: %d\n",
		who, desc, actiontime)
}

func setNewConfig() {
	who, desc, actiontime, err := promptSetNewConfig()
	if err != nil {
		panic(err)
	}

	err = WriteConfigFile(who, desc, actiontime)
	if err != nil {
		panic(err)
	}
}

func promptSetNewConfig() (int64, string, int64, error) {
	who, err := utils.PromptInteger("kto")
	if err != nil {
		return 0, "", 1, err
	}

	desc, err := utils.PromptString("opis")
	if err != nil {
		return 0, "", 1, err
	}

	actiontime, err := utils.PromptInteger("czas")
	if err != nil {
		return 0, "", 1, err
	}

	return who, desc, actiontime, nil
}

func WriteConfigFile(who int64, desc string, actiontime int64) error {
	_, err := os.Create(path.Join(".", utils.DefaultConfigFilename+".toml"))
	if err != nil {
		return err
	}
	viper.SetConfigName(utils.DefaultConfigFilename)
	viper.AddConfigPath(".")
	viper.Set("kto", who)
	viper.Set("opis", desc)
	viper.Set("czas", actiontime)
	return viper.WriteConfig()
}
