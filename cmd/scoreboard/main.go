package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/engine"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/logging"
)

type Item struct {
	Name      string `json:"name"`
	Collected bool   `json:"collected"`
}

type LevelScore struct {
	Level          string `json:"level"`
	CollectedItems []Item `json:"items"`
}

type TeamScore struct {
	Name      string       `json:"name"`
	Score     int          `json:"score"`
	UpdatedAt time.Time    `json:"updatedAt"`
	Levels    []LevelScore `json:"levels"`
}

type Team struct {
	Name          string   `mapstructure:"name"`
	SnapshotsPath []string `mapstructure:"snapshots_path"`
}
type Config struct {
	Listen string `mapstructure:"listen"`
	Teams  []Team `mapstructure:"teams"`
}

func readConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading yaml config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

type snapshot struct {
	path  string
	time  time.Time
	level string
}

func getScoreboard(cfg *Config) ([]TeamScore, error) {
	var scores []TeamScore
	for _, team := range cfg.Teams {
		var teamScore TeamScore
		teamScore.Name = team.Name
		latestSnapshotByLevel := make(map[string]snapshot)

		for _, snapshotPath := range team.SnapshotsPath {
			files, err := os.ReadDir(snapshotPath)
			if err != nil {
				return nil, fmt.Errorf("listing snapshots directory: %w", err)
			}

			for _, f := range files {
				if !f.Type().IsRegular() || !strings.HasPrefix(f.Name(), "snapshot") {
					continue
				}

				parts := strings.Split(f.Name(), "_")
				if len(parts) != 3 {
					return nil, fmt.Errorf("invalid snapshot filename: %s", f.Name())
				}

				level := parts[1]
				dt := parts[2]

				t, err := time.Parse("2006-01-02T15:04:05.999999999", dt)
				if err != nil {
					return nil, fmt.Errorf("parsing snapshot time: %w", err)
				}

				if s, ok := latestSnapshotByLevel[level]; ok && t.Before(s.time) {
					continue
				}

				latestSnapshotByLevel[level] = snapshot{
					path:  path.Join(snapshotPath, f.Name()),
					time:  t,
					level: level,
				}
			}
		}

		var (
			totalScore  int
			lastUpdated time.Time
		)

		for _, s := range latestSnapshotByLevel {
			var e engine.Engine
			f, err := os.Open(s.path)
			if err != nil {
				return nil, fmt.Errorf("reading snapshot file: %w", err)
			}

			if err := json.NewDecoder(f).Decode(&e); err != nil {
				return nil, fmt.Errorf("decoding snapshot: %w", err)
			}

			levelScore := LevelScore{
				Level: s.level,
			}

			for _, it := range e.Items {
				if !it.Important {
					continue
				}
				levelScore.CollectedItems = append(levelScore.CollectedItems, Item{
					Name:      it.Name,
					Collected: it.Collected,
				})

				if it.Collected {
					totalScore++
				}
			}

			if lastUpdated.Before(s.time) {
				lastUpdated = s.time
			}

			teamScore.Levels = append(teamScore.Levels, levelScore)
		}

		teamScore.Score = totalScore
		teamScore.UpdatedAt = lastUpdated
		scores = append(scores, teamScore)
	}
	return scores, nil
}

func main() {
	config := pflag.String("config", "config.yml", "path to the config file")
	pflag.Parse()
	logging.Init()

	cfg, err := readConfig(*config)
	if err != nil {
		logrus.Fatalf("reading config: %v", err)
	}

	logrus.Infof("Parsed config: %+v", cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/scoreboard", func(w http.ResponseWriter, r *http.Request) {
		scores, err := getScoreboard(cfg)
		if err != nil {
			http.Error(w, fmt.Sprintf("getting scoreboard: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(scores); err != nil {
			http.Error(w, fmt.Sprintf("encoding response: %v", err), http.StatusInternalServerError)
			return
		}
	})

	if err := http.ListenAndServe(cfg.Listen, mux); err != nil {
		logrus.Fatalf("serving: %v", err)
	}
}
