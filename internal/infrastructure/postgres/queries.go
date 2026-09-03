package postgres

type OPERATOR string

const (
	// Begin query:
	BeginQuery = "BEGIN;"

	// Commit query:
	CommitQuery = "COMMIT;"

	// # SQL Operator And
	SqlOperatorAnd OPERATOR = " AND "

	// Left join query:
	//
	//  pattern:
	//		1. tablename
	//		2. alias for tablename
	//		3. conditions
	LeftJoinQuery = "LEFT JOIN %s AS %s ON %s"

	// Genearal Create query Returning Id:
	//
	//  pattern:
	//		1. tablename
	//		2. columns
	//		3. placeholders_for_values
	CreateQueryReturningId = "INSERT INTO %s (%s) VALUES %s " +
		"ON CONFLICT DO NOTHING RETURNING id;"

	// Fetch query with where:
	//
	//  pattern:
	//		1. columns
	//		2. tablename
	//		3. where_conditions
	FetchQueryWhere = "SELECT %s FROM %s WHERE %s;"

	// Fetch query with where:
	//
	//  pattern:
	//		1. columns
	//		2. tablename
	//		3. where_conditions
	FetchQueryWhereForUpdate = "SELECT %s FROM %s WHERE %s FOR UPDATE;"

	// // Sum query with joins, where and group:
	// //
	// //  pattern:
	// // 		1. group_name
	// // 		2. column to be summed
	// //		3. tablename
	// //		4. join statements
	// //		5. where_conditions
	// // 		6. group_name
	// SumQueryJoinWhereGroup = "SELECT %s FROM %s %s WHERE %s GROUP BY %s FOR UPDATE;"

	// Genearal Insert query:
	//
	//  pattern:
	//		1. tablename
	//		2. columns
	//		3. placeholders_for_values
	CreateQuery = "INSERT INTO %s (%s) VALUES %s;"

	// Genearal Update query:
	//
	//  pattern:
	//		1. tablename
	//		2. columns
	//		3. where_conditions
	UpdateQueryWhere = "UPDATE %s SET %s WHERE %s;"

	// Genearal Update From query:
	//
	//  pattern:
	//		1. tablename
	//		2. columns
	//		3. from join tablename
	//		4. where_conditions
	UpdateQueryFromWhere = "UPDATE %s SET %s FROM %s WHERE %s;"
)
