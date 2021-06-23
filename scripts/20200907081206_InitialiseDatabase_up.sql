CREATE TABLE `businesses`
(
    `id`              int PRIMARY KEY AUTO_INCREMENT,
    `registered_name` varchar(100),
    `address_id`      int,
    `display_name`    varchar(100),
    `use_preferred`   tinyint(1),
    `tel_number`      varchar(50),
    `website`         varchar(100),
    `business_email`  varchar(50)
);

CREATE TABLE `addresses`
(
    `id`       int PRIMARY KEY AUTO_INCREMENT,
    `value`    varchar(50),
    `street`   varchar(50),
    `suburb`   varchar(50),
    `city`     varchar(50),
    `country`  varchar(50),
    `postcode` varchar(50),
    `state`    varchar(50)
);

CREATE TABLE `business_users`
(
    `id`          int PRIMARY KEY AUTO_INCREMENT,
    `business_id` int,
    `user_id`     int
);

CREATE TABLE `bank_accounts`
(
    `id`             int PRIMARY KEY AUTO_INCREMENT,
    `swift_code`     varchar(50),
    `branch_code`    varchar(50),
    `account_number` varchar(50),
    `tax_number`     varchar(50),
    `branch_id`      int
);

CREATE TABLE `billing_settings`
(
    `id`                 int PRIMARY KEY AUTO_INCREMENT,
    `business_branch_id` int,
    `business_id`        int,
    `price_per_unit`     double,
    `vat_rate`           double,
    `vat_included`       tinyint(1),
    `currency_id`        int
);

CREATE TABLE `currencies`
(
    `id`            int PRIMARY KEY AUTO_INCREMENT,
    `name`          varchar(50),
    `abbreviation`  varchar(10),
    `dollar_index`  double,
    `exchange_rate` double
);

CREATE TABLE `business_branches`
(
    `id`             int PRIMARY KEY AUTO_INCREMENT,
    `business_id`    int,
    `branch_name`    varchar(50),
    `telephone`      varchar(15),
    `website`        varchar(50),
    `business_email` varchar(50)
);

CREATE TABLE `users`
(
    `id`                  int PRIMARY KEY AUTO_INCREMENT,
    `reference_id`        char(36),
    `title`               varchar(100),
    `email`               varchar(60),
    `name`                varchar(100),
    `surname`             varchar(100),
    `tel_number`          varchar(15),
    `username`            varchar(20),
    `access_failed_count` int,
    `lockout_enabled`     tinyint(1),
    `password_hash`       varchar(50),
    `concurrency_stamp`   varchar(50),
    `security_stamp`      varchar(50),
    `password_salt`       varchar(50),
    `accept_terms`        tinyint(1),
    `reset_token`         varchar(100),
    `verification_token`  varchar(100),
    `verification_date`   datetime,
    `password_reset`      datetime,
    `reset_token_expires` datetime,
    `date_created`        datetime,
    `date_updated`        datetime,
    `is_live`             tinyint(1),
    `address_id`          int,
    `business_branch_id`  int
);

CREATE TABLE `user_roles`
(
    `id`      int PRIMARY KEY AUTO_INCREMENT,
    `user_id` int,
    `role_id` int
);

CREATE TABLE `roles`
(
    `id`        int PRIMARY KEY AUTO_INCREMENT,
    `role_name` varchar(50)
);

CREATE TABLE `refresh_tokens`
(
    `id`                bigint PRIMARY KEY AUTO_INCREMENT,
    `user_id`           int,
    `token`             varchar(100),
    `expires`           datetime,
    `is_expired`        tinyint(1),
    `date_created`      datetime,
    `created_by_ip`     varchar(30),
    `revoked`           datetime,
    `revoked_by_ip`     varchar(30),
    `replaced_by_token` varchar(100),
    `is_active`         tinyint(1)
);

CREATE TABLE `user_metrics`
(
    `id`                 int PRIMARY KEY AUTO_INCREMENT,
    `user_id`            int,
    `productivity_index` double,
    `distance_travelled` double
);

CREATE TABLE `drivers`
(
    `id`                   int PRIMARY KEY AUTO_INCREMENT,
    `id_number`            varchar(13),
    `licence_type_id`      int,
    `license_renewal_date` datetime,
    `license_issued`       datetime,
    `user_id`              int
);

CREATE TABLE `lookup_license_types`
(
    `id`   int PRIMARY KEY AUTO_INCREMENT,
    `name` varchar(100)
);

CREATE TABLE `vehicles`
(
    `id`                            int PRIMARY KEY AUTO_INCREMENT,
    `license_plate`                 varchar(50),
    `license_expiry`                varchar(50),
    `licence_disc`                  varchar(50),
    `vehicle_identification_number` varchar(50),
    `model`                         varchar(50),
    `make`                          varchar(50),
    `acquisition_date`              datetime,
    `vehicle_type_id`               int
);

CREATE TABLE `lookup_vehicle_types`
(
    `id`   int PRIMARY KEY AUTO_INCREMENT,
    `name` varchar(50)
);

CREATE TABLE `service_history`
(
    `id`                    int PRIMARY KEY AUTO_INCREMENT,
    `vehicle_id`            int,
    `note`                  varchar(50),
    `next_service_date`     datetime,
    `previous_service_date` datetime,
    `distance_travelled`    bigint
);

CREATE TABLE `shipments`
(
    `id`           bigint PRIMARY KEY AUTO_INCREMENT,
    `guid`         char(36),
    `volume`       double,
    `weight`       double,
    `distance`     double,
    `pickup`       datetime,
    `delivered_on` datetime,
    `departure`    datetime,
    `dispatch_id`  bigint
);

CREATE TABLE `travel_sentiments`
(
    `id`          bigint PRIMARY KEY AUTO_INCREMENT,
    `fuel_cost`   double,
    `fuel_usage`  double,
    `temperature` double,
    `wind_speed`  double,
    `comment`     varchar(50),
    `shipment_id` bigint
);

CREATE TABLE `location_pings`
(
    `id`         bigint PRIMARY KEY AUTO_INCREMENT,
    `vehicle_id` int,
    `latitude`   double,
    `longitude`  double,
    `tracker_id` int
);

CREATE TABLE `trackers`
(
    `id`            int PRIMARY KEY AUTO_INCREMENT,
    `serial_number` varchar(25),
    `vehicle_id`    int
);

CREATE TABLE `invoices`
(
    `id`                 bigint PRIMARY KEY AUTO_INCREMENT,
    `name`               varchar(50),
    `business_branch_id` int,
    `status_id`          int,
    `payment_method_id`  int
);

CREATE TABLE `invoice_items`
(
    `id`          bigint PRIMARY KEY AUTO_INCREMENT,
    `name`        varchar(50),
    `description` varchar(50),
    `cost`        double,
    `amount`      int,
    `discount`    double,
    `invoice_id`  bigint
);

CREATE TABLE `lookup_payment_methods`
(
    `id`   int PRIMARY KEY AUTO_INCREMENT,
    `name` varchar(50)
);

CREATE TABLE `lookup_invoice_statuses`
(
    `id`   int PRIMARY KEY AUTO_INCREMENT,
    `name` varchar(50)
);

CREATE TABLE `clients`
(
    `id`            int PRIMARY KEY AUTO_INCREMENT,
    `business_name` varchar(50),
    `email`         varchar(50),
    `tel`           varchar(50),
    `address_id`    int
);

CREATE TABLE `branch_clients`
(
    `id`                 int PRIMARY KEY AUTO_INCREMENT,
    `business_branch_id` int,
    `client_id`          int
);

CREATE TABLE `client_representatives`
(
    `id`            int PRIMARY KEY AUTO_INCREMENT,
    `name`          varchar(100),
    `surname`       varchar(100),
    `email`         varchar(50),
    `position`      varchar(50),
    `tel_number`    varchar(20),
    `client_id`     int,
    `ghost_user_id` int
);

CREATE TABLE `ghost_users`
(
    `id`           int PRIMARY KEY AUTO_INCREMENT,
    `access_token` varchar(100),
    `display_name` varchar(50)
);

CREATE TABLE `dispatches`
(
    `id`           bigint PRIMARY KEY AUTO_INCREMENT,
    `name`         varchar(100),
    `date_created` datetime,
    `date_updated` datetime,
    `guid`         char(36)
);

CREATE TABLE `app_configurations`
(
    `id`                         int PRIMARY KEY AUTO_INCREMENT,
    `name`                       varchar(100),
    `value`                      varchar(1000),
    `business_id`                int,
    `business_branch_id`         int,
    `app_configuration_group_id` int
);

CREATE TABLE `app_configuration_groups`
(
    `id`   int PRIMARY KEY AUTO_INCREMENT,
    `name` varchar(80)
);

CREATE TABLE `user_positions`
(
    `id`          int PRIMARY KEY AUTO_INCREMENT,
    `user_id`     int,
    `position_id` int
);

CREATE TABLE `positions`
(
    `id`   int PRIMARY KEY AUTO_INCREMENT,
    `name` varchar(100)
);

ALTER TABLE `bank_accounts`
    ADD CONSTRAINT `bank_accounts_branch_id`
        FOREIGN KEY (`branch_id`)
            REFERENCES `business_branches` (`id`);

ALTER TABLE `billing_settings`
    ADD CONSTRAINT `billing_settings_business_branch_id`
        FOREIGN KEY (`business_branch_id`)
            REFERENCES `business_branches` (`id`);

ALTER TABLE `billing_settings`
    ADD CONSTRAINT `billing_settings_currency_id`
        FOREIGN KEY (`currency_id`)
            REFERENCES `currencies` (`id`);

ALTER TABLE `business_branches`
    ADD CONSTRAINT `business_branches_business_id`
        FOREIGN KEY (`business_id`)
            REFERENCES `businesses` (`id`);

ALTER TABLE `users`
    ADD CONSTRAINT `users_business_branch_id`
        FOREIGN KEY (`business_branch_id`)
            REFERENCES `business_branches` (`id`);

ALTER TABLE `users`
    ADD CONSTRAINT `unique_user_email`
        UNIQUE (`email`);

ALTER TABLE `user_roles`
    ADD CONSTRAINT `user_roles_user_id`
        FOREIGN KEY (`user_id`)
            REFERENCES `users` (`id`);

ALTER TABLE `user_roles`
    ADD CONSTRAINT `user_roles_role_id`
        FOREIGN KEY (`role_id`)
            REFERENCES `roles` (`id`);

ALTER TABLE `refresh_tokens`
    ADD CONSTRAINT `refresh_tokens_user_id`
        FOREIGN KEY (`user_id`)
            REFERENCES `users` (`id`);

ALTER TABLE `user_metrics`
    ADD CONSTRAINT `user_metrics_user_id`
        FOREIGN KEY (`user_id`)
            REFERENCES `users` (`id`);

ALTER TABLE `drivers`
    ADD CONSTRAINT `drivers_user_id`
        FOREIGN KEY (`user_id`)
            REFERENCES `users` (`id`);

ALTER TABlE `drivers`
    ADD CONSTRAINT `drivers_license_type_id`
        FOREIGN KEY (`licence_type_id`)
            REFERENCES `lookup_license_types` (`id`);

ALTER TABLE `vehicles`
    ADD CONSTRAINT `vehicles_vehicle_type_id`
        FOREIGN KEY (`vehicle_type_id`)
            REFERENCES `lookup_vehicle_types` (`id`);

ALTER TABLE `service_history`
    ADD CONSTRAINT `service_history_vehicle_id`
        FOREIGN KEY (`vehicle_id`)
            REFERENCES `vehicles` (`id`);

ALTER TABLE `shipments`
    ADD CONSTRAINT `shipments_dispatch_id`
        FOREIGN KEY (`dispatch_id`)
            REFERENCES `dispatches` (`id`);

ALTER TABLE `travel_sentiments`
    ADD CONSTRAINT `travel_sentiments_shipment_id`
        FOREIGN KEY (`shipment_id`)
            REFERENCES `shipments` (`id`);

ALTER TABLE `trackers`
    ADD CONSTRAINT `trackers_vehicle_id`
        FOREIGN KEY (`vehicle_id`)
            REFERENCES `vehicles` (`id`);

ALTER TABLE `location_pings`
    ADD CONSTRAINT `location_pings_vehicle_id`
        FOREIGN KEY (`vehicle_id`)
            REFERENCES `vehicles` (`id`);

ALTER TABLE `location_pings`
    ADD CONSTRAINT `location_pings_tracker_id`
        FOREIGN KEY (`tracker_id`)
            REFERENCES `trackers` (`id`);

ALTER TABLE `invoices`
    ADD CONSTRAINT `invoices_business_branch_id`
        FOREIGN KEY (`business_branch_id`)
            REFERENCES `business_branches` (`id`);

ALTER TABLE `invoices`
    ADD CONSTRAINT `invoices_payment_method_id`
        FOREIGN KEY (`payment_method_id`)
            REFERENCES `lookup_payment_methods` (`id`);

ALTER TABLE `invoices`
    ADD CONSTRAINT `invoices_status_id`
        FOREIGN KEY (`status_id`)
            REFERENCES `lookup_invoice_statuses` (`id`);

ALTER TABLE `invoice_items`
    ADD CONSTRAINT `invoice_items_invoice_id`
        FOREIGN KEY (`invoice_id`)
            REFERENCES `invoices` (`id`);

ALTER TABLE `clients`
    ADD CONSTRAINT `clients_address_id`
        FOREIGN KEY (`address_id`)
            REFERENCES `addresses` (`id`);

ALTER TABLE `branch_clients`
    ADD CONSTRAINT `branch_clients_business_branch_id`
        FOREIGN KEY (`business_branch_id`)
            REFERENCES `business_branches` (`id`);

ALTER TABLE `branch_clients`
    ADD CONSTRAINT `branch_clients_client_id`
        FOREIGN KEY (`client_id`)
            REFERENCES `clients` (`id`);

ALTER TABLE `client_representatives`
    ADD CONSTRAINT `client_representatives_client_id`
        FOREIGN KEY (`client_id`)
            REFERENCES `clients` (`id`);

ALTER TABLE `client_representatives`
    ADD CONSTRAINT `client_representatives_ghost_user_id`
        FOREIGN KEY (`ghost_user_id`)
            REFERENCES `ghost_users` (`id`);

ALTER TABLE `app_configurations`
    ADD CONSTRAINT `app_configurations_business_branch_id`
        FOREIGN KEY (`business_branch_id`)
            REFERENCES `business_branches` (`id`);

ALTER TABLE `app_configurations`
    ADD CONSTRAINT `app_configurations_app_configuration_group_id`
        FOREIGN KEY (`app_configuration_group_id`)
            REFERENCES `app_configuration_groups` (`id`);

ALTER TABLE `user_positions`
    ADD CONSTRAINT `user_positions_position_id`
        FOREIGN KEY (`position_id`)
            REFERENCES `positions` (`id`);

ALTER TABLE `user_positions`
    ADD CONSTRAINT `user_positions_user_id`
        FOREIGN KEY (`user_id`)
            REFERENCES `users` (`id`);

ALTER TABLE `businesses`
    ADD CONSTRAINT `businesses_address_id`
        FOREIGN KEY (`address_id`)
            REFERENCES `addresses` (`id`);

ALTER TABLE `business_users`
    ADD CONSTRAINT `business_users_user_id`
        FOREIGN KEY (`user_id`)
            REFERENCES `users` (`id`);

ALTER TABLE `business_users`
    ADD CONSTRAINT `business_users_business_id`
        FOREIGN KEY (`business_id`)
            REFERENCES `businesses` (`id`);