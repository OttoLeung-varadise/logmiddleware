# logmiddleware
A log middleware,allow you to add request  logs in postgresql.

Only for Gin.

Logs will write into your database, which is power by PostgreSQl 16+.

## Installation
```shell
go get github.com/OttoLeung-varadise/logmiddleware
```

## Before Use
1. You Should make sure your database is power by PostgreSQL 16+.
2. Create a database : **request-log** into your .
3. Run this Schume to create the log's table.
     ```SQL
     -- DROP SCHEMA public;

        CREATE SCHEMA public AUTHORIZATION pg_database_owner;

    -- public.request_logs definition

    -- Drop table

    -- DROP TABLE public.request_logs;

        CREATE TABLE public.request_logs (
            id serial4 NOT NULL,
            request_id varchar(64) NOT NULL,
            "method" varchar(10) NOT NULL,
            "path" varchar(255) NOT NULL,
            "service_name" varchar(255) NULL,
            query_string text NULL,
            status_code int8 NOT NULL,
            remote_ip varchar(45) NOT NULL,
            user_agent text NULL,
            request_time numeric NOT NULL,
            created_at timestamptz DEFAULT now() NOT NULL,
            file_name varchar(255) NULL,
            file_size int8 NULL,
            content_type varchar(45) NULL,
            content_json jsonb NULL,
            CONSTRAINT request_logs_pkey PRIMARY KEY (id)
        );
        CREATE INDEX idx_request_logs_created_at ON public.request_logs USING btree (created_at);
        CREATE INDEX idx_request_logs_path ON public.request_logs USING btree (path);
        CREATE INDEX idx_request_logs_request_id ON public.request_logs USING btree (request_id);
     ```

## How to Use

Add following code in your main.go

```golang
import (
	loggerModel "github.com/OttoLeung-varadise/logmiddleware/model"
    midLogger "github.com/OttoLeung-varadise/logmiddleware/logger"
)


func main() {
    logDB, logErr := loggerModel.InitLogDB()
	if logErr != nil {
		fmt.Printf("log database init fails: %v\n", logErr)
	}
	r := gin.Default()

    if logErr == nil {
		// create gorouties, use the logDB connetion.
		go midLogger.StartLogWriter(logDB)
		// use logger middleware
		r.Use(midLogger.RequestLogMiddleware())
	}

    ....your codes....
}
```

configaretion your env
```shell
    DB_HOST=$(your_database_host)
	DB_PORT=$(your_database_port)
	DB_USER=$(your_databese_user)
	DB_PASSWORD=$(your_database_passwoed)
```