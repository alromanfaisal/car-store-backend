-- migrations/003_seed_cars.sql
INSERT INTO cars (name, description, image_url, price, discount_price, is_new) VALUES
('Toyota Prius', 'Model 2026', '/images/Toyota-prius-2026.png', 5000000, NULL, TRUE),
('Toyota Yaris', 'Model 2026', '/images/Toyota-yaris-2026.png', 3000000, NULL, TRUE),
('Toyota Prius', 'Model 2022', '/images/Toyota-prius-2022.png', 2500000, 2000000, FALSE),
('Toyota Rav4', 'Model 2026', '/images/Toyota-rav4-2026.png', 6000000, NULL, TRUE),
('Toyota Yaris', 'Model 2020', '/images/Toyota-yaris-2020.png', 1800000, 1400000, FALSE),
('Toyota Fortuner', 'Model 2026', '/images/Toyota-fortunar-2026.png', 2500000, NULL, TRUE),
('Product Seven', 'Full page exclusive item.', '/images/product7.png', 3000000, NULL, FALSE),
('Toyota Rav4', 'Model 2020', '/images/Toyota-rav4-2020.png', 2200000, 1800000, FALSE),
('Product Nine', 'Full page exclusive item.', '/images/product9.png', 2700000, NULL, TRUE),
('Product Ten', 'Full page exclusive item.', '/images/product10.png', 3200000, NULL, TRUE),
('Toyota Fielder', 'Model 2020', '/images/Toyota-fielder-2020.png', 3500000, 3000000, FALSE),
('Product Twelve', 'Full page exclusive item.', '/images/product12.png', 4000000, NULL, FALSE),
('Product Thirteen', 'Full page exclusive item.', '/images/product13.png', 4500000, NULL, TRUE),
('Product Fourteen', 'Full page exclusive item.', '/images/product14.png', 5000000, 4500000, FALSE),
('Product Fifteen', 'Full page exclusive item.', '/images/product15.png', 5500000, NULL, TRUE),
('Product Sixteen', 'Full page exclusive item.', '/images/product16.png', 6000000, NULL, TRUE);