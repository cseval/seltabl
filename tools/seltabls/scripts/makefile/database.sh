#!/bin/bash
# file: makefile/database.sh
# url: https://github.com/conneroisu/seltab/tools/seltab-lsp/scripts/makefile/database.sh
# title: Generates the database for the project
# description: This script generates the database for the project
dbs=(
	"master"
)

# for each known database
for db in "${dbs[@]}"; do
	awk 'FNR==1{print ""}1' ./data/"$db"/schemas/*.sql > "./data/$db/combined/schema.sql"
	awk 'FNR==1{print ""}1' ./data/"$db"/seeds/*.sql > "./data/$db/combined/seeds.sql"
	awk 'FNR==1{print ""}1' ./data/"$db"/queries/*.sql > "./data/$db/combined/queries.sql"
done

for db in "${dbs[@]}"; do
	cd "./data/$db" || echo "db $db not found" && exit
	sqlc generate
	cd "../.." || echo "parent folder not found" && exit
	rm "./data/$db/db.go"
done
