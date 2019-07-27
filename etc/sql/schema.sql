BEGIN;


CREATE TABLE "exchange" (
    id   SMALLINT PRIMARY KEY,
    name VARCHAR(20) NOT NULL,

    CONSTRAINT "exchange_to_name_unique" UNIQUE (name)
);

INSERT INTO "exchange" ("id", "name") VALUES (1, 'BINANCE');
INSERT INTO "exchange" ("id", "name") VALUES (2, 'KRAKEN');


CREATE TABLE "direction" (
    id    SMALLINT PRIMARY KEY,
    value VARCHAR(5) NOT NULL,

    CONSTRAINT "direction_to_value_unique" UNIQUE (value)
);

INSERT INTO "direction" ("id", "value") VALUES (1, 'BUY');
INSERT INTO "direction" ("id", "value") VALUES (2, 'SELL');


CREATE TABLE "order_type" (
    id   SMALLINT PRIMARY KEY,
    type VARCHAR(20) NOT NULL,

    CONSTRAINT "order_type_to_type_unique" UNIQUE (type)
);

INSERT INTO "order_type" ("id", "type") VALUES (1, 'MARKET');
INSERT INTO "order_type" ("id", "type") VALUES (2, 'LIMIT');


CREATE TABLE "time_in_force" (
    id   SMALLINT PRIMARY KEY,
    type VARCHAR(5) NOT NULL,

    CONSTRAINT "time_in_force_to_type_unique" UNIQUE (type)
);

INSERT INTO "time_in_force" ("id", "type") VALUES (1, 'FOK');
INSERT INTO "time_in_force" ("id", "type") VALUES (2, 'GTC');


CREATE TABLE "execution_type" (
    id   SMALLINT PRIMARY KEY,
    type VARCHAR(10) NOT NULL,

    CONSTRAINT "execution_type_to_type_unique" UNIQUE (type)
);

INSERT INTO "execution_type" ("id", "type") VALUES (1, 'OPEN');
INSERT INTO "execution_type" ("id", "type") VALUES (2, 'CLOSE');


CREATE TABLE "execution_status" (
    id    SMALLINT PRIMARY KEY,
    value VARCHAR(20) NOT NULL,

    CONSTRAINT "execution_status_to_value_unique" UNIQUE (value)
);

INSERT INTO "execution_status" ("id", "value") VALUES (1, 'CREATED');
INSERT INTO "execution_status" ("id", "value") VALUES (2, 'EXECUTING');
INSERT INTO "execution_status" ("id", "value") VALUES (3, 'SENT');
INSERT INTO "execution_status" ("id", "value") VALUES (4, 'COMPLETED');
INSERT INTO "execution_status" ("id", "value") VALUES (5, 'TIMED_OUT');
INSERT INTO "execution_status" ("id", "value") VALUES (6, 'ERROR');
INSERT INTO "execution_status" ("id", "value") VALUES (7, 'REJECTED');


CREATE TABLE "execution" (
    id                BIGSERIAL PRIMARY KEY,
    exchange_id       SMALLINT NOT NULL,
    instrument_name   VARCHAR(20) NOT NULL,
    direction_id      SMALLINT NOT NULL,
    order_type_id     SMALLINT NOT NULL,
    limit_price       NUMERIC(24,10) DEFAULT NULL,
    amount            NUMERIC(24, 10) NOT NULL,
    status_id         SMALLINT NOT NULL,
    connector_id      SMALLINT DEFAULT NULL,
    execution_type_id SMALLINT NOT NULL,
    execute_till_time TIMESTAMP NOT NULL,
    ref_position_id   TEXT DEFAULT NULL,
    time_in_force_id  SMALLINT NOT NULL,
    update_timestamp  TIMESTAMP NOT NULL,
    account_id        BIGINT NOT NULL,
    api_key           TEXT NOT NULL,
    secret_key        TEXT NOT NULL,
    description       TEXT DEFAULT NULL,
    result_order_id   TEXT DEFAULT NULL,
    finger_print      TEXT NOT NULL,

    CONSTRAINT "execution_fk1" FOREIGN KEY ("exchange_id")       REFERENCES "exchange"         ("id"),
    CONSTRAINT "execution_fk2" FOREIGN KEY ("status_id")         REFERENCES "execution_status" ("id"),
    CONSTRAINT "execution_fk3" FOREIGN KEY ("direction_id")      REFERENCES "direction"        ("id"),
    CONSTRAINT "execution_fk4" FOREIGN KEY ("order_type_id")     REFERENCES "order_type"       ("id"),
    CONSTRAINT "execution_fk5" FOREIGN KEY ("time_in_force_id")  REFERENCES "time_in_force"    ("id"),
    CONSTRAINT "execution_fk6" FOREIGN KEY ("execution_type_id") REFERENCES "execution_type"   ("id"),
    CONSTRAINT unique_finger_print UNIQUE(finger_print)
);


CREATE TABLE "execution_history" (
    id             BIGSERIAL PRIMARY KEY,
    execution_id   BIGINT NOT NULL,
    status_from_id SMALLINT NOT NULL,
    status_to_id   SMALLINT NOT NULL,
    timestamp      TIMESTAMP NOT NULL,
    description    TEXT DEFAULT NULL,

    CONSTRAINT "execution_history_fk1" FOREIGN KEY ("execution_id")   REFERENCES "execution" ("id"),
    CONSTRAINT "execution_history_fk2" FOREIGN KEY ("status_from_id") REFERENCES "execution_status" ("id"),
    CONSTRAINT "execution_history_fk3" FOREIGN KEY ("status_to_id")   REFERENCES "execution_status" ("id")
);


COMMIT;

