package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/logging"
)

//go:embed frontend
var frontendFS embed.FS

type Item struct {
	Name        string    `json:"name"`
	Collected   bool      `json:"collected"`
	CollectedAt time.Time `json:"collectedAt,omitempty"`
}

type LevelScore struct {
	Level string `json:"level"`
	Items []Item `json:"items"`
}

type TeamScore struct {
	Name      string        `json:"name"`
	Score     int           `json:"score"`
	TotalTime time.Duration `json:"totalTime"`
	UpdatedAt time.Time     `json:"updatedAt"`
	Levels    []LevelScore  `json:"levels"`
}

type Team struct {
	Name          string   `mapstructure:"name"`
	SnapshotsPath []string `mapstructure:"snapshots_path"`
}
type Round struct {
	Level string `mapstructure:"level"`
	Start string `mapstructure:"start"`
	start time.Time
}
type Config struct {
	Listen     string  `mapstructure:"listen"`
	Teams      []Team  `mapstructure:"teams"`
	Rounds     []Round `mapstructure:"rounds"`
	SpritesDir string  `mapstructure:"sprites_dir"`
}

type snapshotItem struct {
	Name      string `json:"name"`
	Important bool   `json:"important"`
	Collected bool   `json:"collected"`
}

type snapshotEngine struct {
	Items     []snapshotItem `json:"items"`
	CreatedAt time.Time      `json:"created_at"`
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

	for i, round := range cfg.Rounds {
		t, err := time.Parse("2006-01-02T15:04:05", round.Start)
		if err != nil {
			return nil, fmt.Errorf("parsing round start time: %w", err)
		}
		cfg.Rounds[i].start = t
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

		var teamSnapshots []string
		for _, snapshotPath := range team.SnapshotsPath {
			files, err := os.ReadDir(snapshotPath)
			if err != nil {
				return nil, fmt.Errorf("listing snapshots directory: %w", err)
			}
			for _, f := range files {
				if !f.Type().IsRegular() || !strings.HasPrefix(f.Name(), "snapshot") {
					continue
				}
				teamSnapshots = append(teamSnapshots, path.Join(snapshotPath, f.Name()))
			}
		}

		var (
			totalScore  int
			lastUpdated time.Time
			totalTime   time.Duration
		)

		for _, round := range cfg.Rounds {
			levelSnapshots := lo.Filter(teamSnapshots, func(snapshot string, _ int) bool {
				return strings.Contains(snapshot, round.Level)
			})
			slices.Sort(levelSnapshots)
			if len(levelSnapshots) == 0 {
				continue
			}

			collectedAt := make(map[string]time.Time)

			for _, sp := range levelSnapshots {

				var e snapshotEngine
				f, err := os.Open(sp)
				if err != nil {
					return nil, fmt.Errorf("reading snapshot file: %w", err)
				}

				if err := json.NewDecoder(f).Decode(&e); err != nil {
					return nil, fmt.Errorf("decoding snapshot: %w", err)
				}

				for _, it := range e.Items {
					if !it.Important {
						continue
					}
					if _, ok := collectedAt[it.Name]; !ok {
						// Still show not collected items later.
						collectedAt[it.Name] = time.Time{}
					}
					// Only update collectedAt if it was not collected before.
					if it.Collected && collectedAt[it.Name].IsZero() {
						collectedAt[it.Name] = e.CreatedAt
					}
				}
			}

			var (
				lastCollectedAt time.Time
				levelScore      LevelScore
			)

			for item, ct := range collectedAt {
				levelScore.Items = append(levelScore.Items, Item{
					Name:        item,
					Collected:   !ct.IsZero(),
					CollectedAt: ct,
				})
				if !ct.IsZero() {
					totalScore += 1
				}
				if ct.After(lastCollectedAt) {
					lastCollectedAt = ct
				}
			}
			slices.SortFunc(levelScore.Items, func(a, b Item) int {
				return strings.Compare(a.Name, b.Name)
			})

			levelScore.Level = round.Level
			teamScore.Levels = append(teamScore.Levels, levelScore)
			totalTime += lastCollectedAt.Sub(round.start)
			if lastCollectedAt.After(lastUpdated) {
				lastUpdated = lastCollectedAt
			}
		}

		teamScore.TotalTime = totalTime
		teamScore.Score = totalScore
		teamScore.UpdatedAt = lastUpdated
		scores = append(scores, teamScore)
	}

	slices.SortFunc(scores, func(a, b TeamScore) int {
		if a.Score == b.Score {
			return int(a.TotalTime.Milliseconds() - b.TotalTime.Milliseconds())
		}
		return b.Score - a.Score
	})
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
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFileFS(w, r, frontendFS, "frontend/index.html")
			return
		}
		http.FileServer(http.FS(frontendFS)).ServeHTTP(w, r)
	})
	if cfg.SpritesDir != "" {
		mux.Handle("/sprites/", http.StripPrefix("/sprites/", http.FileServer(http.Dir(cfg.SpritesDir))))
	}
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
