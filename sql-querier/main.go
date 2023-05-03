package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
)

func main() {
	var (
		host, user, password, name string
		port                       int
		init                       bool
	)

	// Parse command-line flags
	flag.StringVar(&host, "db-host", "localhost", "Database host")
	flag.IntVar(&port, "db-port", 5432, "Database port")
	flag.StringVar(&user, "db-user", "", "Database user")
	flag.StringVar(&password, "db-password", "", "Database password")
	flag.StringVar(&name, "db-name", "northwind", "Database name")
	flag.BoolVar(&init, "init", false, "Initialize database")
	flag.Parse()

	querier, err := connect(host, port, user, password, name)
	failOnError(err, "Failed to connect to database")
	defer querier.Close()
	fmt.Println("Connected to Northwind database")

	if init {
		err = querier.Init()
		failOnError(err, "Failed to initialize database")
		fmt.Println("Initialized Northwind database")
	}

	fmt.Println("\nGetting all categories...")
	categories, err := querier.GetAllCategories()
	failOnError(err, "Failed to get categories")
	printMany("Categories", categories)

	fmt.Println("\nGetting all shippers...")
	shippers, err := querier.GetAllShippers()
	failOnError(err, "Failed to get shippers")
	printMany("Shippers", shippers)

	fmt.Println("\nGetting all 'beverage' products...")
	products, err := querier.GetProductsByCategory(1) // 1 - Beverages
	failOnError(err, "Failed to get products")
	printMany("'Beverage' products", products)

	fmt.Println("\nGetting 'Outback Lager' product...")
	product, err := querier.GetProductByID(70) // 70 - Outback Lager
	failOnError(err, "Failed to get product")
	printSingle("'Outback Lager' product", product)

	fmt.Println("\nGetting all products with 'Lager' in their name...")
	products, err = querier.GetProductsByName("Lager")
	failOnError(err, "Failed to get products")
	printMany("'Lager' products", products)

	fmt.Println("\nGetting all products in the price range of $20.00 to $50.00...")
	products, err = querier.GetProductsByPriceRange(20.00, 50.00)
	failOnError(err, "Failed to get products")
	printMany("Products in the price range of $20.00 to $50.00", products)

	fmt.Println("\nGetting the average price of all products...")
	avgPrice, err := querier.GetAverageProductPrice()
	failOnError(err, "Failed to get average price")
	fmt.Printf("Average price: $%.2f\n", avgPrice)

	fmt.Println("\nGetting top 5 most expensive products...")
	products, err = querier.GetTopExpensiveProducts(5)
	failOnError(err, "Failed to get products")
	printMany("Top 5 most expensive products", products)

	fmt.Println("\nGetting products networth...")
	networth, err := querier.GetProductsNetworth()
	failOnError(err, "Failed to get products")
	printSingle("Products networth", networth)

	fmt.Println("\nGetting all countries of customers...")
	countries, err := querier.GetAllCountriesOfCustomers()
	failOnError(err, "Failed to get countries")
	printMany("Countries of customers", countries)

	fmt.Println("\nGetting top 5 countries of customers...")
	topCountries, err := querier.GetTopCountriesOfCustomers(5)
	failOnError(err, "Failed to get countries")
	printMany("Top 5 countries of customers", topCountries)

	fmt.Println("\nGetting customers from London...")
	customersFromCity, err := querier.GetCustomersFromCity("London")
	failOnError(err, "Failed to get customers")
	printMany("Customers from London", customersFromCity)

	fmt.Println("\nGetting customers named Mario...")
	customersByName, err := querier.GetCustomersByName("Mario")
	failOnError(err, "Failed to get customers")
	printMany("Customers named Mario", customersByName)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func connect(host string, port int, user string, password string, dbname string) (*Querier, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open connection to %q: %w", connStr, err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("ping %q: %w", connStr, err)
	}

	return NewQuerier(db), nil
}

func printSingle[T any](name string, entry T) {
	fmt.Printf("%s: %+v\n", name, entry)
}

func printMany[T any](name string, entries []T) {
	fmt.Printf("%s:\n", name)
	for i, entity := range entries {
		fmt.Printf("\t%d: %+v\n", i, entity)
	}
}
