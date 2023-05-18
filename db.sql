create table items (
	id	bigserial not null primary key,
	Chrt_id      int,
	Track_number varchar(128),	
	Price      	 int,    
	Rid        	 varchar(256), 
	Name       	 varchar(128), 
	Sale       	 int,    
	Size       	 varchar(128), 
	TotalPrice 	 int,    
	Nm_id        int,    
	Brand      	 varchar(128),
	Status		 int
);
create table delivery (
	id 	bigserial not null primary key,
	Name 	varchar(128),
	Phone 	varchar(128),
	Zip 	varchar(128),
	City 	varchar(128),
	Address varchar(128),
	Region 	varchar(128),
	Email 	varchar(128)
);

create table payment (
	id	bigserial not null primary key,
	Transaction   varchar(256),
	Requist_id	  varchar(128),
	Currency      varchar(128), 
	Provider      varchar(128),
	Amount        int,
	Payment_dt    int,  
	Bank          varchar(128),
	Delivery_cost int,
	Goods_total   int,
	Custom_fee	  int
);

create table "orders" (
	id	bigserial not null primary key,
	OrderUID          varchar(128),  
	Entry             varchar(128), 
	InternalSignature varchar(128),  
	delivery_id_fk		 bigserial,
	payment_id_fk        bigserial,
	Locale            varchar(128), 
	CustomerID        varchar(128), 
	TrackNumber       varchar(128),  
	DeliveryService   varchar(128), 
	Shardkey          varchar(128),  
	SmID              int,
	totalprice        int
);

create table "order_items" (
	id	bigserial not null primary key, 
	order_id_fk        bigserial,
	item_id_fk         bigserial
);

create table "cache" (
	id	bigserial not null primary key, 
	order_id	int8, 
	app_key        varchar(128)
);

ALTER TABLE public.orders ADD CONSTRAINT payment_id_fkey FOREIGN KEY (payment_id_fk) REFERENCES public.payment(id) on update no action on delete no action not valid;
ALTER TABLE public.order_items ADD CONSTRAINT order_id_fkey FOREIGN KEY (order_id_fk) REFERENCES public.orders(id) match simple on update no action on delete no action not valid;
ALTER TABLE public.order_items ADD CONSTRAINT item_id_fkey FOREIGN KEY (item_id_fk) REFERENCES public.items(id) match simple on update no action on delete no action not valid;
ALTER TABLE public.orders ADD CONSTRAINT delivery_id_fk FOREIGN KEY (delivery_id_fk) REFERENCES public.delivery(id) on update no action on delete no action not valid;
