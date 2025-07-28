#!/bin/bash

MONGO_URL="mongodb://appuser:apppassword@127.0.0.1:27017/pointdb?authSource=pointdb"

mongosh "$MONGO_URL" --eval "db.customers.deleteMany({})"
