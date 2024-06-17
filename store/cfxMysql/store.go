package cfxMysql

import (
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
)

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Name     string `yaml:"name"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"database"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func NewDB(config *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Database.User, config.Database.Password, config.Database.Host, config.Database.Port, config.Database.Name)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func InitDB(db *gorm.DB) error {
	err := db.AutoMigrate(&Block{}, &Address{}, &TransactionDetail{}, &Log{}, &Hash{}, &Erc20Transfer{})
	if err != nil {
		return err
	}
	return nil
}

func StoreBlockTransactionsAndLogs(db *gorm.DB, blockDataMySQL BlockDataMySQL) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&blockDataMySQL.Block).Error; err != nil {
			return fmt.Errorf("failed to create block: %w", err)
		}
		if len(blockDataMySQL.TransactionDetails) == 0 {
			return nil
		}
		if err := tx.Create(&blockDataMySQL.TransactionDetails).Error; err != nil {
			return fmt.Errorf("failed to create transaction details: %w", err)
		}
		if len(blockDataMySQL.Logs) == 0 {
			return nil
		}
		if err := tx.Create(&blockDataMySQL.Logs).Error; err != nil {
			return fmt.Errorf("failed to create logs: %w", err)
		}
		if err := tx.Create(&blockDataMySQL.Erc20Transfers).Error; err != nil {
			return fmt.Errorf("failed to create logs: %w", err)
		}
		return nil
	})
}

func FindOrCreateAddress(db *gorm.DB, addressStr string) (Address, error) {
	var address Address
	if err := db.Where("address = ?", addressStr).First(&address).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Address not found, create a new address
			address = Address{Address: addressStr}
			if err = db.Create(&address).Error; err != nil {
				return address, fmt.Errorf("failed to create address: %w", err)
			}
		} else {
			return address, fmt.Errorf("failed to query address: %w", err)
		}
	}
	return address, nil
}

func FindOrCreateHash(db *gorm.DB, hashStr string) (Hash, error) {
	var hash Hash
	if err := db.Where("hash = ?", hashStr).First(&hash).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Hash not found, create a new hash
			hash = Hash{Hash: hashStr}
			if err = db.Create(&hash).Error; err != nil {
				return hash, fmt.Errorf("failed to create hash: %w", err)
			}
		} else {
			return hash, fmt.Errorf("failed to query hash: %w", err)
		}
	}
	return hash, nil
}

func FindErc20(db *gorm.DB, addressStr string) (bool, error) {
	var erc20 Erc20
	if err := db.Where("address = ?", addressStr).First(&erc20).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func CreateErc20(db *gorm.DB, address string) error {
	erc20 := Erc20{Address: address}
	if err := db.Create(&erc20).Error; err != nil {
		return fmt.Errorf("failed to create ERC20 address: %w", err)
	}
	return nil
}

func getLatestBlock(db *gorm.DB) (Block, error) {
	var block Block
	if err := db.Order("block_number desc").First(&block).Error; err != nil {
		return block, fmt.Errorf("failed to get latest block: %w", err)
	}
	return block, nil
}

func checkBlockExists(db *gorm.DB, blockNumber uint) (bool, error) {
	var block Block
	if err := db.First(&block, "block_number = ?", blockNumber).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if block exists: %w", err)
	}
	return true, nil
}

func GetInitBlockNumber(db *gorm.DB, inputBlockNumber uint64) (uint64, error) {
	exists, err := checkBlockExists(db, uint(inputBlockNumber))
	if err != nil {
		fmt.Printf("failed to check if block exists: %v\n", err)
		return 0, err
	}
	if exists {
		latestBlock, err := getLatestBlock(db)
		if err != nil {
			fmt.Printf("failed to get latest block: %v\n", err)
			return 0, err
		} else {
			return uint64(latestBlock.BlockNumber + 1), nil
		}
	} else {
		return inputBlockNumber, nil
	}
}

func Start() *gorm.DB {
	// Load configuration
	config, err := LoadConfig("./store/cfxMysql/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := NewDB(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = InitDB(db)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	return db

}
