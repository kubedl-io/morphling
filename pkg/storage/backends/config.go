/*
Copyright 2021 The Alibaba Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package backends

import (
	"fmt"
	"github.com/alibaba/morphling/pkg/controllers/consts"
	"os"
	"strconv"
)

const (
	EnvDBHost     = "MYSQL_HOST"
	EnvDBPort     = "MYSQL_PORT"
	EnvDBDatabase = "MYSQL_DB_NAME"
	EnvDBUser     = "MYSQL_USER"
	EnvDBPassword = "MYSQL_PASSWORD"
	EnvLogMode    = "MYSQL_LOGMODE"
)

func GetMysqlDBSource() (dbSource, logMode string, err error) {
	host := GetEnvOrDefault(EnvDBHost, consts.DefaultMorphlingMySqlServiceName)
	port, err := strconv.Atoi(GetEnvOrDefault(EnvDBPort, consts.DefaultMorphlingMySqlServicePort))
	if err != nil {
		return "", "", err
	}

	db := GetEnvOrDefault(EnvDBDatabase, "morphling")
	user := GetEnvOrDefault(EnvDBUser, "root")
	password := GetEnvOrDefault(EnvDBPassword, "morphling")

	// Expected: "root:morphling@tcp(morphling-mysql:3306)/morphling?timeout=5s"
	dbSource = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=5s", user, password, host, port, db)
	logMode = GetEnvOrDefault(EnvLogMode, "no")

	return dbSource, logMode, nil
}

func GetEnvOrDefault(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
