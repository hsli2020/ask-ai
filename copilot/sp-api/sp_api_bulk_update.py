import requests
import json
import time
from typing import Dict, List, Any

class AmazonSPAPI:
    def __init__(self, access_token: str, refresh_token: str, client_id: str, client_secret: str, seller_id: str, region: str = "na"):
        self.access_token = access_token
        self.refresh_token = refresh_token
        self.client_id = client_id
        self.client_secret = client_secret
        self.seller_id = seller_id
        self.region = region
        self.base_url = f"https://sellingpartnerapi-{region}.amazon.com"
        
    def get_headers(self) -> Dict[str, str]:
        """获取API调用所需的headers"""
        return {
            "Authorization": f"Bearer {self.access_token}",
            "x-amz-access-token": self.access_token,
            "Content-Type": "application/json"
        }
    
    def create_feed_document(self, content_type: str = "application/json") -> Dict[str, Any]:
        """创建Feed文档"""
        url = f"{self.base_url}/feeds/2021-06-30/documents"
        payload = {
            "contentType": content_type
        }
        
        response = requests.post(url, json=payload, headers=self.get_headers())
        return response.json()
    
    def upload_feed_data(self, upload_url: str, data: str):
        """上传Feed数据到S3"""
        headers = {"Content-Type": "application/json"}
        response = requests.put(upload_url, data=data, headers=headers)
        return response.status_code == 200
    
    def create_feed(self, feed_type: str, marketplace_ids: List[str], document_id: str) -> Dict[str, Any]:
        """创建Feed"""
        url = f"{self.base_url}/feeds/2021-06-30/feeds"
        payload = {
            "feedType": feed_type,
            "marketplaceIds": marketplace_ids,
            "inputFeedDocument": document_id
        }
        
        response = requests.post(url, json=payload, headers=self.get_headers())
        return response.json()
    
    def get_feed_status(self, feed_id: str) -> Dict[str, Any]:
        """获取Feed状态"""
        url = f"{self.base_url}/feeds/2021-06-30/feeds/{feed_id}"
        response = requests.get(url, headers=self.get_headers())
        return response.json()
    
    def bulk_update_products(self, products_data: List[Dict[str, Any]], marketplace_ids: List[str]):
        """批量更新商品信息"""
        # 1. 创建JSON格式的Feed数据
        json_data = self.create_json_listings_feed(products_data)
        
        # 2. 创建Feed文档
        feed_doc = self.create_feed_document("application/json")
        if "uploadDestination" not in feed_doc:
            raise Exception("Failed to create feed document")
            
        # 3. 上传数据
        upload_success = self.upload_feed_data(
            feed_doc["uploadDestination"]["url"], 
            json_data
        )
        if not upload_success:
            raise Exception("Failed to upload feed data")
            
        # 4. 创建Feed
        feed_result = self.create_feed(
            "JSON_LISTINGS_FEED",
            marketplace_ids,
            feed_doc["feedDocumentId"]
        )
        
        return feed_result
    
    def create_json_listings_feed(self, products: List[Dict[str, Any]]) -> str:
        """创建JSON_LISTINGS_FEED格式的数据"""
        feed_data = {
            "header": {
                "sellerId": self.seller_id,
                "version": "2.0",
                "issueLocale": "en_US"
            },
            "messages": []
        }
        
        for i, product in enumerate(products, 1):
            message = {
                "messageId": i,
                "sku": product["sku"],
                "operationType": "UPDATE",
                "productType": "PRODUCT",
                "requirements": "LISTING",
                "attributes": self._build_attributes(product)
            }
            feed_data["messages"].append(message)
        
        return json.dumps(feed_data, indent=2)
    
    def _build_attributes(self, product: Dict[str, Any]) -> Dict[str, Any]:
        """构建商品属性"""
        attributes = {
            "condition_type": [{"value": "new_new"}]
        }
        
        # 构建价格信息
        purchasable_offer = {
            "currency": "USD"
        }
        
        if "price" in product:
            purchasable_offer["our_price"] = [{
                "schedule": [{"value_with_tax": product["price"]}]
            }]
        
        if "minimum_seller_allowed_price" in product:
            purchasable_offer["minimum_seller_allowed_price"] = [{
                "schedule": [{"value_with_tax": product["minimum_seller_allowed_price"]}]
            }]
            
        if "maximum_seller_allowed_price" in product:
            purchasable_offer["maximum_seller_allowed_price"] = [{
                "schedule": [{"value_with_tax": product["maximum_seller_allowed_price"]}]
            }]
            
        if "business_price" in product:
            purchasable_offer["business_price"] = [{
                "schedule": [{"value_with_tax": product["business_price"]}]
            }]
        
        # 构建分层价格
        if "quantity_discounts" in product and product["quantity_discounts"]:
            quantity_discounts = []
            discount_type = product.get("quantity_price_type", "FIXED_AMOUNT").upper()
            
            purchasable_offer["quantity_discount_type"] = [{"value": discount_type}]
            
            for i, discount in enumerate(product["quantity_discounts"], 1):
                discount_item = {
                    "quantity_tier": i,
                    "quantity_discount_type": discount_type,
                    "quantity_lower_bound": discount["quantity_lower_bound"]
                }
                
                if discount_type == "FIXED_AMOUNT":
                    # 计算固定折扣金额
                    base_price = product.get("price", 0)
                    discount_price = discount.get("quantity_price", base_price)
                    discount_amount = base_price - discount_price
                    discount_item["discount_amount"] = max(0, discount_amount)
                else:  # PERCENT_OFF
                    base_price = product.get("price", 0)
                    discount_price = discount.get("quantity_price", base_price)
                    if base_price > 0:
                        discount_percent = ((base_price - discount_price) / base_price) * 100
                        discount_item["discount_percent"] = max(0, discount_percent)
                    else:
                        discount_item["discount_percent"] = 0
                
                quantity_discounts.append(discount_item)
            
            purchasable_offer["quantity_discount"] = quantity_discounts
        
        attributes["purchasable_offer"] = [purchasable_offer]
        
        # 构建库存信息
        if "quantity" in product or "handling_time" in product:
            fulfillment_availability = {
                "fulfillment_channel_code": "DEFAULT"
            }
            
            if "quantity" in product:
                fulfillment_availability["quantity"] = product["quantity"]
                
            if "handling_time" in product:
                fulfillment_availability["lead_time_to_ship_max_days"] = product["handling_time"]
            
            attributes["fulfillment_availability"] = [fulfillment_availability]
        
        return attributes

# 使用示例
if __name__ == "__main__":
    # 初始化API客户端
    api = AmazonSPAPI(
        access_token="your_access_token",
        refresh_token="your_refresh_token", 
        client_id="your_client_id",
        client_secret="your_client_secret",
        seller_id="your_seller_id"
    )
    
    # 准备要更新的商品数据
    products_to_update = [
        {
            "sku": "YOUR-SKU-001",
            "price": 29.99,
            "minimum_seller_allowed_price": 25.00,
            "maximum_seller_allowed_price": 35.00,
            "quantity": 100,
            "handling_time": 2,
            "business_price": 27.99,
            "quantity_price_type": "FIXED_AMOUNT",
            "quantity_discounts": [
                {"quantity_lower_bound": 10, "quantity_price": 28.99},
                {"quantity_lower_bound": 50, "quantity_price": 27.99},
                {"quantity_lower_bound": 100, "quantity_price": 26.99}
            ]
        },
        {
            "sku": "YOUR-SKU-002", 
            "price": 19.99,
            "minimum_seller_allowed_price": 15.00,
            "maximum_seller_allowed_price": 25.00,
            "quantity": 200,
            "handling_time": 1,
            "business_price": 18.99,
            "quantity_price_type": "PERCENT_OFF",
            "quantity_discounts": [
                {"quantity_lower_bound": 20, "quantity_price": 18.99},
                {"quantity_lower_bound": 100, "quantity_price": 17.99},
                {"quantity_lower_bound": 200, "quantity_price": 16.99}
            ]
        }
    ]
    
    # 执行批量更新
    marketplace_ids = ["ATVPDKIKX0DER"]  # 美国站点
    
    try:
        result = api.bulk_update_products(products_to_update, marketplace_ids)
        print(f"Feed创建成功，Feed ID: {result.get('feedId')}")
        
        # 检查Feed状态
        feed_id = result.get('feedId')
        if feed_id:
            time.sleep(5)  # 等待处理
            status = api.get_feed_status(feed_id)
            print(f"Feed状态: {status.get('processingStatus')}")
            
    except Exception as e:
        print(f"更新失败: {str(e)}")