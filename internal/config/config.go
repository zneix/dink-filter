package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	RoutePrefix  string                        `json:"routePrefix"`
	BindAddress  string                        `json:"bindAddress"`
	Password     string                        `json:"password"`
	GlobalFilter DestinationFilter             `json:"globalFilter"`
	Destinations map[string]*DestinationFilter `json:"destinations"`
}

type DestinationFilter struct {
	EnableLoot               bool            `json:"enableLoot"`
	EnableKillCount          bool            `json:"enableKillCount"`
	EnableKillCountRegular   bool            `json:"enableKillCountRegular"`
	EnableKillCountPBs       bool            `json:"enableKillCountPBs"`
	LootThreshold            *int            `json:"lootThreshold,omitempty"`
	DefaultKillCountInterval *int            `json:"defaultKillCountInterval,omitempty"`
	KillCountIntervals       *map[string]int `json:"killCountIntervals,omitempty"`
	// TODO: Implement "always fire if killCount == X" option
}

func New() *Config {
	data, err := os.ReadFile("config.json")
	if err != nil {
		panic(fmt.Errorf("failed to read config file: %w", err))
	}

	cfg := new(Config)
	if err = json.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Errorf("failed to unmarshal config file: %w", err))
	}

	// Check validity of necessary config values
	if len(cfg.RoutePrefix) > 0 && !strings.HasPrefix(cfg.RoutePrefix, "/") {
		panic("error initializing config: \"routePrefix\" must start with '/'")
	}

	if cfg.GlobalFilter.LootThreshold == nil || cfg.GlobalFilter.DefaultKillCountInterval == nil {
		panic("error initializing config: global filter values should be specified")
	}

	return cfg
}

func (c *Config) GetLootTreshold(url string) int {
	// 1. Check for url-specific loot treshold
	urlFilter := c.Destinations[url]
	if urlFilter != nil && urlFilter.LootThreshold != nil {
		return *urlFilter.LootThreshold
	}

	// 2. As fallback, use global default value
	return *c.GlobalFilter.LootThreshold
}

func (c *Config) GetKillCountInterval(url, bossName string) int {
	// url-specific configuration
	urlFilter := c.Destinations[url]

	// 1. Check for boss-specific intervals for this url
	if urlFilter != nil && urlFilter.KillCountIntervals != nil {
		if bossSpecificInterval, exists := (*urlFilter.KillCountIntervals)[bossName]; exists {
			return bossSpecificInterval
		}
	}

	// 2. Check for global boss-specific intervals
	if c.GlobalFilter.KillCountIntervals != nil {
		if globalBossSpecifcInterval, exists := (*c.GlobalFilter.KillCountIntervals)[bossName]; exists {
			return globalBossSpecifcInterval
		}
	}

	// 3. With no boss-specific intervals, check for url-specific default interval
	if urlFilter != nil && urlFilter.DefaultKillCountInterval != nil {
		return *urlFilter.DefaultKillCountInterval
	}

	// 4. As fallback, use global default interval
	return *c.GlobalFilter.DefaultKillCountInterval
}
