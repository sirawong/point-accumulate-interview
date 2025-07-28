db = db.getSiblingDB('pointdb');

db.createUser({
    user: 'appuser',
    pwd: 'apppassword',
    roles: [
        {
            role: 'readWrite',
            db: 'pointdb'
        }
    ]
});

db.createCollection('customers');

db.customers.createIndex({"customer_id": 1}, {unique: true});

db.createCollection('rules');
db.rules.createIndex({
        "status": 1,
        "conditions.branch_id": 1,
        "conditions.category_ids": 1,
    },
    {
        name: "status_1_conditions.branch_id_1_conditions.category_ids_1"
    }
);

if (db.rules.countDocuments() === 0) {
    print("Collection is empty. Inserting initial data...");

    const rulesToInsert = [
        {
            name: "BEVERAGE (BR3444) - 1 point per 100 THB",
            rule_type: "RATIO",
            conditions: {
                min_amount: Decimal128("100.00"),
                branch_id: "BR3444",
                category_ids: ["CT1001"]
            },
            reward: { value: NumberLong(1), ratio_unit: 100.0 },
            status: "ACTIVE",
        },
        {
            name: "BEVERAGE (BR3456) - Bonus 2 points over 200 THB",
            rule_type: "FIXED_POINT",
            conditions: {
                min_amount: Decimal128("200.01"),
                branch_id: "BR3456",
                category_ids: ["CT1001"]
            },
            reward: { value: NumberLong(2) },
            status: "ACTIVE",
        },
        {
            name: "BEVERAGE (BR3458) - 5% points",
            rule_type: "PERCENTAGE",
            conditions: {
                min_amount: Decimal128("0.01"),
                branch_id: "BR3458",
                category_ids: ["CT1001"]
            },
            reward: { value: NumberLong(5) },
            status: "ACTIVE",
        },
        {
            name: "ELECTRONICS (BR3444) - 10% points over 500 THB",
            rule_type: "PERCENTAGE",
            conditions: {
                min_amount: Decimal128("500.01"),
                branch_id: "BR3444",
                category_ids: ["CT1002"]
            },
            reward: { value: NumberLong(10) },
            status: "ACTIVE",
        },
        {
            name: "ELECTRONICS (BR3456) - 5 points per 500 THB",
            rule_type: "RATIO",
            conditions: {
                min_amount: Decimal128("500.00"),
                branch_id: "BR3456",
                category_ids: ["CT1002"]
            },
            reward: { value: NumberLong(5), ratio_unit: 500.0 },
            status: "ACTIVE",
        },
        {
            name: "FOOD (BR3444) - Bonus 10 points over 500 THB",
            rule_type: "FIXED_POINT",
            conditions: {
                min_amount: Decimal128("500.01"),
                branch_id: "BR3444",
                category_ids: ["CT1003"]
            },
            reward: { value: NumberLong(10) },
            status: "ACTIVE",
        },
        {
            name: "FOOD (BR3456) - 2% points over 100 THB",
            rule_type: "PERCENTAGE",
            conditions: {
                min_amount: Decimal128("100.01"),
                branch_id: "BR3456",
                category_ids: ["CT1003"]
            },
            reward: { value: NumberLong(2) },
            status: "ACTIVE",
        },
        {
            name: "FOOD (BR1111) - 2% points",
            rule_type: "PERCENTAGE",
            conditions: {
                min_amount: Decimal128("0.01"),
                branch_id: "BR1111",
                category_ids: ["CT1003"]
            },
            reward: { value: NumberLong(2) },
            status: "ACTIVE",
        }
    ];

    db.rules.insertMany(rulesToInsert);
    print(`Inserted ${rulesToInsert.length} initial rules.`);

} else {
    print("Collection already has data. Skipping insertion.");
}

print('Database initialization completed!');


