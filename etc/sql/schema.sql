BEGIN;

CREATE TABLE "exchange" (
    id SMALLSERIAL PRIMARY KEY,
    name VARCHAR(20) NOT NULL,

    CONSTRAINT "exchange_to_name_unique" UNIQUE (name)
);

INSERT INTO "exchange" ("name") VALUES ('BINANCE');
INSERT INTO "exchange" ("name") VALUES ('KRAKEN');


CREATE TABLE "direction" (
    id SMALLSERIAL PRIMARY KEY,
    value VARCHAR(5) NOT NULL,

    CONSTRAINT "direction_to_value_unique" UNIQUE (value)
);

INSERT INTO "direction" ("value") VALUES ('BUY');
INSERT INTO "direction" ("value") VALUES ('SELL');


CREATE TABLE "order_type" (
    id SMALLSERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL,

    CONSTRAINT "order_type_to_type_unique" UNIQUE (type)
);

INSERT INTO "order_type" ("type") VALUES ('MARKET');
INSERT INTO "order_type" ("type") VALUES ('LIMIT');


CREATE TABLE "time_in_force" (
    id SMALLSERIAL PRIMARY KEY,
    type VARCHAR(5) NOT NULL,

    CONSTRAINT "time_in_force_to_type_unique" UNIQUE (type)
);

INSERT INTO "time_in_force" ("type") VALUES ('FOK');


CREATE TABLE "execution_type" (
    id SMALLSERIAL PRIMARY KEY,
    type VARCHAR(10) NOT NULL,

    CONSTRAINT "execution_type_to_type_unique" UNIQUE (type)
);

INSERT INTO "execution_type" ("type") VALUES ('OPEN');
INSERT INTO "execution_type" ("type") VALUES ('CLOSE');


CREATE TABLE "execution_status" (
    id SMALLSERIAL PRIMARY KEY,
    value VARCHAR(20) NOT NULL,

    CONSTRAINT "execution_status_to_value_unique" UNIQUE (value)
);

INSERT INTO "execution_status" ("value") VALUES ('CREATED');
INSERT INTO "execution_status" ("value") VALUES ('EXECUTING');
INSERT INTO "execution_status" ("value") VALUES ('COMPLETED');
INSERT INTO "execution_status" ("value") VALUES ('TIMED_OUT');
INSERT INTO "execution_status" ("value") VALUES ('ERROR');
INSERT INTO "execution_status" ("value") VALUES ('REJECTED');


COMMIT;

