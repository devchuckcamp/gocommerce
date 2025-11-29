package migrations

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// ProductSeed generates mock product data based on the catalog.Product schema.
var ProductSeed = Seed{
	Name:        "product_seeder",
	Description: "Seeds the products table with realistic mock product data",
	Run:         seedProducts,
}

// seedProducts inserts mock product data matching the catalog.Product schema:
// ID, SKU, Name, Description, BrandID, CategoryID, BasePrice, Status, Images, Attributes, CreatedAt, UpdatedAt
func seedProducts(ctx context.Context, exec Executor) error {
	now := time.Now()

	// Define realistic product data
	products := []struct {
		id          string
		sku         string
		name        string
		description string
		brandID     string
		categoryID  string
		price       int64  // in cents
		currency    string
		status      string
		images      string // JSON array as string
		attributes  string // JSON object as string
	}{
		{
			id:          "prod-laptop-mbp16-001",
			sku:         "MBP16-M3MAX-32-1TB",
			name:        "MacBook Pro 16\" M3 Max",
			description: "Professional laptop with Apple M3 Max chip, 32GB unified memory, 1TB SSD storage. Features stunning Liquid Retina XDR display with ProMotion technology.",
			brandID:     "brand-apple",
			categoryID:  "cat-computers",
			price:       349900,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/mbp16-1.jpg","https://example.com/images/mbp16-2.jpg"]`,
			attributes:  `{"processor":"M3 Max","ram":"32GB","storage":"1TB","screen":"16-inch","color":"Space Black"}`,
		},
		{
			id:          "prod-laptop-dell-xps15",
			sku:         "XPS15-I9-32-1TB",
			name:        "Dell XPS 15",
			description: "Premium Windows laptop featuring Intel Core i9 processor, 32GB DDR5 RAM, 1TB PCIe SSD, and stunning 4K OLED InfinityEdge display.",
			brandID:     "brand-dell",
			categoryID:  "cat-computers",
			price:       279900,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/xps15-1.jpg","https://example.com/images/xps15-2.jpg"]`,
			attributes:  `{"processor":"Intel i9","ram":"32GB","storage":"1TB","screen":"15.6-inch","color":"Platinum Silver"}`,
		},
		{
			id:          "prod-laptop-lenovo-x1",
			sku:         "X1CARBON-I7-16-512",
			name:        "Lenovo ThinkPad X1 Carbon Gen 11",
			description: "Ultra-portable business laptop with Intel Core i7, 16GB RAM, 512GB SSD. Legendary ThinkPad keyboard and military-grade durability.",
			brandID:     "brand-lenovo",
			categoryID:  "cat-computers",
			price:       189900,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/x1carbon-1.jpg"]`,
			attributes:  `{"processor":"Intel i7","ram":"16GB","storage":"512GB","screen":"14-inch","color":"Black"}`,
		},
		{
			id:          "prod-phone-iphone15pro",
			sku:         "IPHONE15PRO-256-TIT",
			name:        "iPhone 15 Pro Max 256GB",
			description: "Latest iPhone with A17 Pro chip, titanium design, advanced camera system with 5x optical zoom. Available in natural titanium finish.",
			brandID:     "brand-apple",
			categoryID:  "cat-electronics",
			price:       119900,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/iphone15pro-1.jpg","https://example.com/images/iphone15pro-2.jpg","https://example.com/images/iphone15pro-3.jpg"]`,
			attributes:  `{"storage":"256GB","color":"Natural Titanium","display":"6.7-inch","chip":"A17 Pro"}`,
		},
		{
			id:          "prod-tablet-ipad-pro",
			sku:         "IPAD-PRO-M2-256",
			name:        "iPad Pro 12.9\" M2",
			description: "Professional tablet with M2 chip, 256GB storage, Liquid Retina XDR display with ProMotion. Compatible with Apple Pencil (2nd gen) and Magic Keyboard.",
			brandID:     "brand-apple",
			categoryID:  "cat-electronics",
			price:       109900,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/ipadpro-1.jpg"]`,
			attributes:  `{"chip":"M2","storage":"256GB","display":"12.9-inch","color":"Space Gray"}`,
		},
		{
			id:          "prod-mouse-mx-master-3s",
			sku:         "MXMASTER3S-BLK",
			name:        "Logitech MX Master 3S",
			description: "Premium wireless mouse with 8K DPI sensor, quiet clicks, ergonomic design. Features MagSpeed scrolling and multi-device connectivity.",
			brandID:     "brand-logitech",
			categoryID:  "cat-accessories",
			price:       9999,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/mxmaster3s-1.jpg"]`,
			attributes:  `{"connectivity":"Bluetooth + USB","dpi":"8000","buttons":"7","color":"Black"}`,
		},
		{
			id:          "prod-keyboard-mx-keys",
			sku:         "MXKEYS-MINI-BLK",
			name:        "Logitech MX Keys Mini",
			description: "Compact wireless illuminated keyboard with smart backlighting, perfect key spacing, and multi-device support. Type on multiple devices seamlessly.",
			brandID:     "brand-logitech",
			categoryID:  "cat-accessories",
			price:       9999,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/mxkeys-1.jpg"]`,
			attributes:  `{"layout":"Compact","connectivity":"Bluetooth + USB","backlight":"Yes","color":"Black"}`,
		},
		{
			id:          "prod-monitor-lg-ultrawide",
			sku:         "LG-34WK95U-5K",
			name:        "LG UltraWide 34\" 5K Monitor",
			description: "34-inch curved ultrawide monitor with 5K resolution (5120x2160), Thunderbolt 4 connectivity, 98% DCI-P3 color gamut. Perfect for creative professionals.",
			brandID:     "brand-samsung",
			categoryID:  "cat-electronics",
			price:       149900,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/lg-ultrawide-1.jpg"]`,
			attributes:  `{"size":"34-inch","resolution":"5K","aspect_ratio":"21:9","panel":"Nano IPS"}`,
		},
		{
			id:          "prod-headphones-sony-xm5",
			sku:         "WH1000XM5-BLK",
			name:        "Sony WH-1000XM5",
			description: "Industry-leading noise cancelling wireless headphones with 30-hour battery life, exceptional sound quality, and multipoint connection.",
			brandID:     "brand-sony",
			categoryID:  "cat-audio",
			price:       39999,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/sony-xm5-1.jpg","https://example.com/images/sony-xm5-2.jpg"]`,
			attributes:  `{"type":"Over-ear","connectivity":"Bluetooth","battery":"30 hours","color":"Black"}`,
		},
		{
			id:          "prod-headphones-bose-qc45",
			sku:         "BOSE-QC45-WHT",
			name:        "Bose QuietComfort 45",
			description: "Premium noise cancelling headphones with legendary comfort, exceptional audio performance, and up to 24 hours of battery life.",
			brandID:     "brand-bose",
			categoryID:  "cat-audio",
			price:       32999,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/bose-qc45-1.jpg"]`,
			attributes:  `{"type":"Over-ear","connectivity":"Bluetooth","battery":"24 hours","color":"White Smoke"}`,
		},
		{
			id:          "prod-webcam-logitech-brio",
			sku:         "BRIO-4K-BLK",
			name:        "Logitech Brio 4K Webcam",
			description: "Professional 4K Ultra HD webcam with HDR, autofocus, and 5x digital zoom. Perfect for video conferencing and streaming.",
			brandID:     "brand-logitech",
			categoryID:  "cat-accessories",
			price:       19999,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/brio-1.jpg"]`,
			attributes:  `{"resolution":"4K UHD","framerate":"30fps","fov":"90 degrees","color":"Black"}`,
		},
		{
			id:          "prod-ssd-samsung-990pro",
			sku:         "990PRO-2TB",
			name:        "Samsung 990 PRO 2TB NVMe SSD",
			description: "High-performance PCIe 4.0 NVMe SSD with sequential read speeds up to 7,450 MB/s. Ideal for gaming, content creation, and heavy workloads.",
			brandID:     "brand-samsung",
			categoryID:  "cat-storage",
			price:       24999,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/990pro-1.jpg"]`,
			attributes:  `{"capacity":"2TB","interface":"PCIe 4.0 x4","form_factor":"M.2 2280","read_speed":"7450 MB/s"}`,
		},
		{
			id:          "prod-ssd-samsung-t7",
			sku:         "T7-PORTABLE-1TB",
			name:        "Samsung T7 Portable SSD 1TB",
			description: "Compact and durable portable SSD with transfer speeds up to 1,050 MB/s. Password protection and AES 256-bit hardware encryption.",
			brandID:     "brand-samsung",
			categoryID:  "cat-storage",
			price:       12999,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/t7-1.jpg"]`,
			attributes:  `{"capacity":"1TB","interface":"USB 3.2 Gen 2","speed":"1050 MB/s","color":"Metallic Red"}`,
		},
		{
			id:          "prod-router-mesh-wifi6e",
			sku:         "MESH-WIFI6E-3PACK",
			name:        "Mesh WiFi 6E System (3-Pack)",
			description: "Whole-home mesh WiFi 6E system covering up to 6,000 sq ft. Lightning-fast speeds, seamless roaming, and advanced security features.",
			brandID:     "brand-hp",
			categoryID:  "cat-networking",
			price:       49999,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/mesh-wifi-1.jpg"]`,
			attributes:  `{"standard":"WiFi 6E","coverage":"6000 sq ft","nodes":"3","max_speed":"6 Gbps"}`,
		},
		{
			id:          "prod-desk-standing",
			sku:         "DESK-STAND-ELEC-60",
			name:        "Electric Standing Desk 60\"",
			description: "Motorized height-adjustable standing desk with memory presets, anti-collision technology, and sturdy steel frame. Supports up to 275 lbs.",
			brandID:     "brand-hp",
			categoryID:  "cat-furniture",
			price:       59999,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/standing-desk-1.jpg"]`,
			attributes:  `{"width":"60 inches","height_range":"28-48 inches","motor":"Dual","color":"Black Oak"}`,
		},
		{
			id:          "prod-chair-ergonomic",
			sku:         "CHAIR-ERGO-MESH",
			name:        "Ergonomic Mesh Office Chair",
			description: "Premium ergonomic office chair with adjustable lumbar support, breathable mesh back, 4D armrests, and smooth-rolling casters.",
			brandID:     "brand-hp",
			categoryID:  "cat-furniture",
			price:       44999,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/ergo-chair-1.jpg"]`,
			attributes:  `{"material":"Mesh","lumbar_support":"Adjustable","armrests":"4D","weight_capacity":"300 lbs"}`,
		},
		{
			id:          "prod-dock-thunderbolt4",
			sku:         "DOCK-TB4-12PORT",
			name:        "Thunderbolt 4 Docking Station",
			description: "Universal Thunderbolt 4 dock with 12 ports including dual 4K display support, 100W power delivery, and Gigabit Ethernet.",
			brandID:     "brand-dell",
			categoryID:  "cat-accessories",
			price:       29999,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/tb4-dock-1.jpg"]`,
			attributes:  `{"ports":"12","displays":"Dual 4K","power_delivery":"100W","ethernet":"Gigabit"}`,
		},
		{
			id:          "prod-keyboard-mech-rgb",
			sku:         "MECH-KB-RGB-CHERRY",
			name:        "Mechanical Gaming Keyboard",
			description: "Premium mechanical keyboard with Cherry MX switches, per-key RGB lighting, aluminum frame, and programmable macros.",
			brandID:     "brand-logitech",
			categoryID:  "cat-accessories",
			price:       17999,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/mech-kb-1.jpg"]`,
			attributes:  `{"switches":"Cherry MX Red","lighting":"RGB","layout":"Full-size","material":"Aluminum"}`,
		},
		{
			id:          "prod-printer-laser-color",
			sku:         "PRINTER-LASER-CLR",
			name:        "Color Laser Printer",
			description: "Fast color laser printer with wireless connectivity, automatic duplex printing, and 250-sheet paper capacity. Perfect for home office.",
			brandID:     "brand-hp",
			categoryID:  "cat-office",
			price:       39999,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/laser-printer-1.jpg"]`,
			attributes:  `{"type":"Laser","color":"Yes","speed":"30 ppm","connectivity":"WiFi + Ethernet"}`,
		},
		{
			id:          "prod-monitor-gaming-240hz",
			sku:         "MONITOR-GAME-27-240",
			name:        "27\" Gaming Monitor 240Hz",
			description: "High-refresh rate gaming monitor with 1ms response time, G-Sync/FreeSync support, and vibrant IPS panel.",
			brandID:     "brand-samsung",
			categoryID:  "cat-electronics",
			price:       49999,
			currency:    "USD",
			status:      "active",
			images:      `["https://example.com/images/gaming-monitor-1.jpg"]`,
			attributes:  `{"size":"27-inch","refresh_rate":"240Hz","response_time":"1ms","panel":"IPS"}`,
		},
		// Draft/Discontinued products for testing different statuses
		{
			id:          "prod-phone-legacy",
			sku:         "PHONE-OLD-001",
			name:        "Legacy Smartphone Model",
			description: "Previous generation smartphone. No longer in production.",
			brandID:     "brand-samsung",
			categoryID:  "cat-electronics",
			price:       39999,
			currency:    "USD",
			status:      "discontinued",
			images:      `[]`,
			attributes:  `{"status":"end_of_life"}`,
		},
		{
			id:          "prod-device-prototype",
			sku:         "PROTO-DEVICE-001",
			name:        "Prototype Smart Device",
			description: "Unreleased product still in development. Not available for purchase.",
			brandID:     "brand-apple",
			categoryID:  "cat-electronics",
			price:       99999,
			currency:    "USD",
			status:      "draft",
			images:      `[]`,
			attributes:  `{"stage":"prototype"}`,
		},
	}

	// Insert products
	for _, p := range products {
		query := `
			INSERT INTO products (
				id, sku, name, description, brand_id, category_id,
				base_price_amount, base_price_currency, status,
				images, attributes, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			ON CONFLICT (id) DO NOTHING
		`

		err := exec.Exec(ctx, query,
			p.id, p.sku, p.name, p.description, p.brandID, p.categoryID,
			p.price, p.currency, p.status,
			p.images, p.attributes, now, now,
		)

		if err != nil {
			return fmt.Errorf("failed to insert product %s: %w", p.name, err)
		}
	}

	return nil
}

// CategorySeed generates mock category data.
var CategorySeed = Seed{
	Name:        "category_seeder",
	Description: "Seeds the categories table with product categories",
	Run:         seedCategories,
}

func seedCategories(ctx context.Context, exec Executor) error {
	now := time.Now()

	categories := []struct {
		id           string
		parentID     *string
		name         string
		slug         string
		description  string
		imageURL     string
		isActive     bool
		displayOrder int
	}{
		{
			id:           "cat-electronics",
			parentID:     nil,
			name:         "Electronics",
			slug:         "electronics",
			description:  "Consumer electronics and gadgets",
			imageURL:     "https://example.com/categories/electronics.jpg",
			isActive:     true,
			displayOrder: 1,
		},
		{
			id:           "cat-computers",
			parentID:     strPtr("cat-electronics"),
			name:         "Computers & Laptops",
			slug:         "computers-laptops",
			description:  "Desktop computers, laptops, and workstations",
			imageURL:     "https://example.com/categories/computers.jpg",
			isActive:     true,
			displayOrder: 1,
		},
		{
			id:           "cat-accessories",
			parentID:     nil,
			name:         "Accessories",
			slug:         "accessories",
			description:  "Computer and electronic accessories",
			imageURL:     "https://example.com/categories/accessories.jpg",
			isActive:     true,
			displayOrder: 2,
		},
		{
			id:           "cat-audio",
			parentID:     strPtr("cat-electronics"),
			name:         "Audio & Headphones",
			slug:         "audio-headphones",
			description:  "Headphones, speakers, and audio equipment",
			imageURL:     "https://example.com/categories/audio.jpg",
			isActive:     true,
			displayOrder: 2,
		},
		{
			id:           "cat-storage",
			parentID:     strPtr("cat-computers"),
			name:         "Storage & Drives",
			slug:         "storage-drives",
			description:  "SSDs, HDDs, and external storage",
			imageURL:     "https://example.com/categories/storage.jpg",
			isActive:     true,
			displayOrder: 3,
		},
		{
			id:           "cat-networking",
			parentID:     nil,
			name:         "Networking",
			slug:         "networking",
			description:  "Routers, switches, and networking equipment",
			imageURL:     "https://example.com/categories/networking.jpg",
			isActive:     true,
			displayOrder: 3,
		},
		{
			id:           "cat-office",
			parentID:     nil,
			name:         "Office Supplies",
			slug:         "office-supplies",
			description:  "Printers, scanners, and office equipment",
			imageURL:     "https://example.com/categories/office.jpg",
			isActive:     true,
			displayOrder: 4,
		},
		{
			id:           "cat-furniture",
			parentID:     nil,
			name:         "Furniture",
			slug:         "furniture",
			description:  "Office desks, chairs, and furniture",
			imageURL:     "https://example.com/categories/furniture.jpg",
			isActive:     true,
			displayOrder: 5,
		},
	}

	for _, c := range categories {
		query := `
			INSERT INTO categories (
				id, parent_id, name, slug, description,
				image_url, is_active, display_order,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (id) DO NOTHING
		`

		err := exec.Exec(ctx, query,
			c.id, c.parentID, c.name, c.slug, c.description,
			c.imageURL, c.isActive, c.displayOrder,
			now, now,
		)

		if err != nil {
			return fmt.Errorf("failed to insert category %s: %w", c.name, err)
		}
	}

	return nil
}

// BrandSeed generates mock brand data.
var BrandSeed = Seed{
	Name:        "brand_seeder",
	Description: "Seeds the brands table with product brands",
	Run:         seedBrands,
}

func seedBrands(ctx context.Context, exec Executor) error {
	now := time.Now()

	brands := []struct {
		id          string
		name        string
		slug        string
		description string
		logoURL     string
		isActive    bool
	}{
		{
			id:          "brand-apple",
			name:        "Apple",
			slug:        "apple",
			description: "American technology company specializing in consumer electronics and software",
			logoURL:     "https://example.com/brands/apple-logo.png",
			isActive:    true,
		},
		{
			id:          "brand-dell",
			name:        "Dell",
			slug:        "dell",
			description: "Global technology leader providing comprehensive solutions",
			logoURL:     "https://example.com/brands/dell-logo.png",
			isActive:    true,
		},
		{
			id:          "brand-lenovo",
			name:        "Lenovo",
			slug:        "lenovo",
			description: "World's leading PC company and emerging PC Plus company",
			logoURL:     "https://example.com/brands/lenovo-logo.png",
			isActive:    true,
		},
		{
			id:          "brand-hp",
			name:        "HP",
			slug:        "hp",
			description: "Technology company providing personal computing and printing solutions",
			logoURL:     "https://example.com/brands/hp-logo.png",
			isActive:    true,
		},
		{
			id:          "brand-samsung",
			name:        "Samsung",
			slug:        "samsung",
			description: "Global leader in technology, opening new possibilities for people everywhere",
			logoURL:     "https://example.com/brands/samsung-logo.png",
			isActive:    true,
		},
		{
			id:          "brand-logitech",
			name:        "Logitech",
			slug:        "logitech",
			description: "Swiss-American manufacturer of computer peripherals and software",
			logoURL:     "https://example.com/brands/logitech-logo.png",
			isActive:    true,
		},
		{
			id:          "brand-sony",
			name:        "Sony",
			slug:        "sony",
			description: "Japanese multinational conglomerate specializing in electronics and entertainment",
			logoURL:     "https://example.com/brands/sony-logo.png",
			isActive:    true,
		},
		{
			id:          "brand-bose",
			name:        "Bose",
			slug:        "bose",
			description: "American manufacturing company specializing in audio equipment",
			logoURL:     "https://example.com/brands/bose-logo.png",
			isActive:    true,
		},
	}

	for _, b := range brands {
		query := `
			INSERT INTO brands (
				id, name, slug, description,
				logo_url, is_active,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO NOTHING
		`

		err := exec.Exec(ctx, query,
			b.id, b.name, b.slug, b.description,
			b.logoURL, b.isActive,
			now, now,
		)

		if err != nil {
			return fmt.Errorf("failed to insert brand %s: %w", b.name, err)
		}
	}

	return nil
}

// RandomProductSeed generates random products for load testing.
var RandomProductSeed = Seed{
	Name:        "random_product_seeder",
	Description: "Seeds the products table with randomly generated products for load testing",
	Run:         seedRandomProducts,
}

func seedRandomProducts(ctx context.Context, exec Executor) error {
	now := time.Now()
	rand.Seed(time.Now().UnixNano())

	categories := []string{"cat-electronics", "cat-computers", "cat-accessories", "cat-audio", "cat-storage"}
	brands := []string{"brand-apple", "brand-dell", "brand-samsung", "brand-logitech", "brand-sony"}
	adjectives := []string{"Premium", "Professional", "Ultimate", "Essential", "Advanced", "Deluxe"}
	productTypes := []string{"Device", "Gadget", "Tool", "Kit", "System", "Solution"}
	statuses := []string{"active", "active", "active", "draft"}

	for i := 1; i <= 50; i++ {
		id := fmt.Sprintf("prod-random-%03d", i)
		sku := fmt.Sprintf("RND-%03d-%d", i, rand.Intn(9999))
		name := fmt.Sprintf("%s %s %s %d",
			adjectives[rand.Intn(len(adjectives))],
			productTypes[rand.Intn(len(productTypes))],
			"Pro",
			rand.Intn(100),
		)
		description := fmt.Sprintf("A high-quality product designed for professional use. Model %d with advanced features.", i)
		brandID := brands[rand.Intn(len(brands))]
		categoryID := categories[rand.Intn(len(categories))]
		price := int64(rand.Intn(100000) + 1000) // $10 to $1000
		status := statuses[rand.Intn(len(statuses))]

		query := `
			INSERT INTO products (
				id, sku, name, description, brand_id, category_id,
				base_price_amount, base_price_currency, status,
				images, attributes, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			ON CONFLICT (id) DO NOTHING
		`

		err := exec.Exec(ctx, query,
			id, sku, name, description, brandID, categoryID,
			price, "USD", status,
			`[]`, `{}`, now, now,
		)

		if err != nil {
			return fmt.Errorf("failed to insert random product %s: %w", name, err)
		}
	}

	return nil
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}

// AllSeeds contains all available seeders.
var AllSeeds = []Seed{
	BrandSeed,
	CategorySeed,
	ProductSeed,
	RandomProductSeed,
}
