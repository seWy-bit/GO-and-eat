CREATE TABLE IF NOT EXISTS orders (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    restaurant_id TEXT NOT NULL,
    status TEXT NOT NULL,
    total_amount BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE RESTRICT
);


CREATE TABLE IF NOT EXISTS order_items (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    menu_item_id TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    price BIGINT NOT NULL,

    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (menu_item_id) REFERENCES menu_items(id) ON DELETE RESTRICT
);

-- Индекс для поиска заказов по пользователю
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);

-- Индекс для поиска заказов по ресторану
CREATE INDEX IF NOT EXISTS idx_order_restaurant_id ON orders(restaurant_id);

-- Индекс для поиска по статусу
CREATE INDEX IF NOT EXISTS idx_order_status ON orders(status);

-- Индекс для поиска позиций заказа
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);

-- Индекс для поиска позиций по блюду
CREATE INDEX IF NOT EXISTS idx_order_items_menu_item_id ON order_items(menu_item_id);

COMMENT ON TABLE orders IS 'Заказы пользователей';
COMMENT ON COLUMN orders.status IS 'Статус заказа: created, confirmed, cancelled';
COMMENT ON COLUMN orders.total_amount IS 'Общая сумма в копейках';

COMMENT ON TABLE order_items IS 'Позиции в заказе';
COMMENT ON COLUMN order_items.price IS 'Цена на момент заказа (фиксируется, чтобы изменения цен не влияли на старые заказы)';