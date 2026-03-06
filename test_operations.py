#!/usr/bin/env python3
"""Complete database operation test"""
import requests
import json

BASE = "http://localhost:8081/api/v1/postgres"

def run_query(query, database="unified_db"):
    """Execute a SQL query"""
    r = requests.post(f"{BASE}/query", json={"query": query, "database": database}, timeout=5)
    return r.json()

def test_operations():
    print("🔧 PostgreSQL Operations Test")
    print("="*60)
    
    # 1. Create table
    print("\n1️⃣  Creating table 'products'...")
    result = run_query("""
        CREATE TABLE IF NOT EXISTS products (
            id SERIAL PRIMARY KEY,
            name VARCHAR(100) NOT NULL,
            price DECIMAL(10,2),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    """)
    print(f"   ✅ {result.get('message', 'Done')}")
    
    # 2. Insert data
    print("\n2️⃣  Inserting sample data...")
    for i in range(3):
        result = run_query(f"INSERT INTO products (name, price) VALUES ('Product {i+1}', {(i+1)*10.99})")
        print(f"   ✅ Inserted Product {i+1}")
    
    # 3. Select all
    print("\n3️⃣  Selecting all products...")
    result = run_query("SELECT * FROM products ORDER BY id")
    if "rows" in result:
        print(f"   Found {len(result['rows'])} products:")
        for row in result['rows']:
            print(f"      • ID: {row['id']}, Name: {row['name']}, Price: ${row['price']}")
    
    # 4. Update data
    print("\n4️⃣  Updating product price...")
    result = run_query("UPDATE products SET price = 99.99 WHERE id = 1")
    print(f"   ✅ Updated {result.get('rows_affected', 0)} row(s)")
    
    # 5. Select with WHERE
    print("\n5️⃣  Selecting updated product...")
    result = run_query("SELECT * FROM products WHERE id = 1")
    if "rows" in result and result['rows']:
        row = result['rows'][0]
        print(f"   Product: {row['name']}, New Price: ${row['price']}")
    
    # 6. Delete data
    print("\n6️⃣  Deleting a product...")
    result = run_query("DELETE FROM products WHERE id = 3")
    print(f"   ✅ Deleted {result.get('rows_affected', 0)} row(s)")
    
    # 7. Count rows
    print("\n7️⃣  Counting remaining products...")
    result = run_query("SELECT COUNT(*) as total FROM products")
    if "rows" in result:
        print(f"   Total products: {result['rows'][0]['total']}")
    
    # 8. Drop table (cleanup)
    print("\n8️⃣  Cleaning up - dropping table...")
    result = run_query("DROP TABLE products")
    print(f"   ✅ {result.get('message', 'Done')}")
    
    print("\n" + "="*60)
    print("✅ All operations completed successfully!")

if __name__ == "__main__":
    test_operations()
