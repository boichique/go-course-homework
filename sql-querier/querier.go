package main

import (
	"database/sql"
	_ "embed" // necessary to use go:embed directive

	_ "github.com/lib/pq" // load "postgres" driver
)

//go:embed northwind.sql
var initScript string

type Querier struct {
	db *sql.DB
}

func NewQuerier(db *sql.DB) *Querier {
	return &Querier{db: db}
}

func (q *Querier) Close() error {
	return q.db.Close()
}

func (q *Querier) Init() error {
	_, err := q.db.Exec(initScript)
	return err
}

// GetAllCategories returns all categories from the categories table
func (q *Querier) GetAllCategories() ([]*Category, error) {
	rows, err := q.db.Query("SELECT category_id, category_name, description FROM categories;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*Category
	for rows.Next() {
		var category Category
		if err = rows.Scan(&category.ID, &category.Name, &category.Description); err != nil {
			return nil, err
		}
		categories = append(categories, &category)
	}
	return categories, nil
}

// GetAllShippers returns all shippers from the shippers table
func (q *Querier) GetAllShippers() ([]*Shipper, error) {
	rows, err := q.db.Query("SELECT shipper_id, company_name, phone FROM shippers;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shippers []*Shipper
	for rows.Next() {
		var shipper Shipper
		if err := rows.Scan(&shipper.ID, &shipper.CompanyName, &shipper.Phone); err != nil {
			return nil, err
		}
		shippers = append(shippers, &shipper)
	}
	return shippers, nil
}

// GetProductsByCategory returns all products from the products table that belong to the given category
func (q *Querier) GetProductsByCategory(categoryID int) ([]*Product, error) {
	rows, err := q.db.Query("SELECT product_id, product_name, supplier_id, category_id, quantity_per_unit, unit_price, units_in_stock, units_on_order, reorder_level, discontinued FROM products WHERE category_id = $1;", categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products, err := scanProducts(rows)
	if err != nil {
		return nil, err
	}
	return products, nil
}

// GetProductByID returns the product with the given ID from the products table
func (q *Querier) GetProductByID(productID int) (*Product, error) {
	row := q.db.QueryRow("SELECT product_id, product_name, supplier_id, category_id, quantity_per_unit, unit_price, units_in_stock, units_on_order, reorder_level, discontinued FROM products WHERE product_id = $1;", productID)

	product, err := scanProduct(row)
	if err != nil {
		return nil, err
	}

	return product, nil
}

// GetProductsByName returns all products from the products table whose name contains the given substring
func (q *Querier) GetProductsByName(substring string) ([]*Product, error) {
	rows, err := q.db.Query("SELECT product_id, product_name, supplier_id, category_id, quantity_per_unit, unit_price, units_in_stock, units_on_order, reorder_level, discontinued FROM products WHERE product_name LIKE '%' || $1 || '%';", substring)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products, err := scanProducts(rows)
	if err != nil {
		return nil, err
	}
	return products, nil
}

// GetProductsByPriceRange returns all products from the products table whose price is between the given values (inclusive)
func (q *Querier) GetProductsByPriceRange(from, to float64) ([]*Product, error) {
	rows, err := q.db.Query("SELECT product_id, product_name, supplier_id, category_id, quantity_per_unit, unit_price, units_in_stock, units_on_order, reorder_level, discontinued FROM products WHERE unit_price BETWEEN $1 AND $2;", from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products, err := scanProducts(rows)
	if err != nil {
		return nil, err
	}
	return products, nil
}

// GetAverageProductPrice returns the average price of all products
func (q *Querier) GetAverageProductPrice() (float64, error) {
	row := q.db.QueryRow("SELECT AVG(product_id) FROM products;")

	var average float64
	err := row.Scan(&average)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return -1, err
	}
	return average, nil
}

// GetTopExpensiveProducts returns the top N most expensive products
func (q *Querier) GetTopExpensiveProducts(top int) ([]*Product, error) {
	rows, err := q.db.Query("SELECT product_id, product_name, supplier_id, category_id, quantity_per_unit, unit_price price, units_in_stock, units_on_order, reorder_level, discontinued FROM products ORDER BY price DESC LIMIT $1;", top)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products, err := scanProducts(rows)
	if err != nil {
		return nil, err
	}

	return products, nil
}

// GetProductsNetworth returns the sum of the price of all products in stock (unit_price * units_in_stock)
func (q *Querier) GetProductsNetworth() (float64, error) {
	row := q.db.QueryRow("SELECT SUM(unit_price * units_in_stock) FROM products;")

	var sum float64
	err := row.Scan(&sum)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return -1, err
	}
	return sum, nil
}

// GetAllCountriesOfCustomers returns all distinct countries from the customers table
func (q *Querier) GetAllCountriesOfCustomers() ([]string, error) {
	rows, err := q.db.Query("SELECT DISTINCT(country) c FROM customers ORDER BY c;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var countries []string
	for rows.Next() {
		var country string
		if err := rows.Scan(&country); err != nil {
			return nil, err
		}
		countries = append(countries, country)
	}
	return countries, nil
}

func (q *Querier) GetTopCountriesOfCustomers(top int) ([]CountryTimes, error) {
	rows, err := q.db.Query("SELECT country, COUNT(customer_id) c FROM customers GROUP BY country ORDER BY c DESC LIMIT $1;", top)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var countries []CountryTimes
	for rows.Next() {
		var country CountryTimes
		if err := rows.Scan(&country.Country, &country.Times); err != nil {
			return nil, err
		}
		countries = append(countries, country)
	}
	return countries, nil
}

func (q *Querier) GetCustomersFromCity(city string) ([]*Customer, error) {
	rows, err := q.db.Query("SELECT customer_id, company_name, contact_name, contact_title, address, city, COALESCE(region, 'None') region, postal_code, country, phone, COALESCE(fax, 'None') fax FROM customers WHERE city LIKE $1;", city)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	customers, err := scanCustomers(rows)
	if err != nil {
		return nil, err
	}
	return customers, nil
}

func (q *Querier) GetCustomersByName(name string) ([]*Customer, error) {
	rows, err := q.db.Query("SELECT customer_id, company_name, contact_name, contact_title, address, city, COALESCE(region, 'None') region, postal_code, country, phone, COALESCE(fax, 'None') fax FROM customers WHERE contact_name LIKE $1 || '%';", name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	customers, err := scanCustomers(rows)
	if err != nil {
		return nil, err
	}
	return customers, nil
}

func scanProduct(row *sql.Row) (*Product, error) {
	var product Product
	err := row.Scan(
		&product.ID,
		&product.Name,
		&product.SupplierID,
		&product.CategoryID,
		&product.QuantityPerUnit,
		&product.UnitPrice,
		&product.UnitsInStock,
		&product.UnitsOnOrder,
		&product.ReorderLevel,
		&product.Discontinued,
	)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func scanProducts(rows *sql.Rows) ([]*Product, error) {
	var products []*Product
	for rows.Next() {
		var product Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.SupplierID,
			&product.CategoryID,
			&product.QuantityPerUnit,
			&product.UnitPrice,
			&product.UnitsInStock,
			&product.UnitsOnOrder,
			&product.ReorderLevel,
			&product.Discontinued,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, &product)
	}
	return products, nil
}

func scanCustomers(rows *sql.Rows) ([]*Customer, error) {
	var customers []*Customer
	for rows.Next() {
		var customer Customer
		err := rows.Scan(
			&customer.ID,
			&customer.Name,
			&customer.ContactName,
			&customer.ContactTitle,
			&customer.Address,
			&customer.City,
			&customer.Region,
			&customer.PostalCode,
			&customer.Country,
			&customer.Phone,
			&customer.Fax,
		)
		if err != nil {
			return nil, err
		}
		customers = append(customers, &customer)
	}
	return customers, nil
}

type Category struct {
	ID          int
	Name        string
	Description string
}

type Shipper struct {
	ID          int
	CompanyName string
	Phone       string
}

type Product struct {
	ID              int
	Name            string
	SupplierID      int
	CategoryID      int
	QuantityPerUnit string
	UnitPrice       float64
	UnitsInStock    int
	UnitsOnOrder    int
	ReorderLevel    int
	Discontinued    bool
}

type CountryTimes struct {
	Country string
	Times   int
}

type Customer struct {
	ID           string
	Name         string
	ContactName  string
	ContactTitle string
	Address      string
	City         string
	Region       string
	PostalCode   string
	Country      string
	Phone        string
	Fax          string
}
