---
type: Knowledge
title: Customer Orders
description: One row per completed customer order across all channels.
tags: [sales, orders, revenue]
---
# Schema

| Column | Type | Description |
| --- | --- | --- |
| `order_id` | STRING | Globally unique order identifier. |

# Joins

Joined with the customer concept on `customer_id`.
