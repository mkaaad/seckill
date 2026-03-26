# 发布订单
## 请求路径：POST /seckill
## 请求参数：application/json
|名称           |类型   |必选   |说明   |
| :------------:| :----:|:----: |:----: |
|product_id     |int    |是     |商品ID |
|strat_time     |time   |是     |开始时间|
|end_time       |time   |是     |结束时间|
|price          |float  |是     |商品价格|
|stock          |int    |是     |库存   |

# 秒杀下单
## 请求路径：GET /order
## 请求参数：query
|名称       |位置       |类型       |必选   |说明   |
| :----:    |:----:     |:----:     |:----: |:----: |
|product_id |query      |int        |是     |商品id |
|user_id    |query      |int        |是     |用户id |

# 订单状态查询
## 请求路径：GET /order/search
## 请求参数：query
|名称       |位置       |类型       |必选   |说明   |
| :----:    |:----:     |:----:     |:----: |:----: |
|user_id    |query      |int        |是     |用户id |
