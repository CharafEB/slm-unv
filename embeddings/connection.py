import redis.asyncio as redis 
import os
from dotenv import load_dotenv

#redis infection dependency
async def redis_connection():

    load_dotenv("../.env")

    redis_conn = redis.Redis(
    host=os.getenv("REDISADD_PY"),
    port=int(os.getenv("REDISPORT")),
    username=os.getenv("REDISUSERNAME"),
    password=os.getenv("REDISPASSWORD"),
    max_connections=20,
    decode_responses=True
)
    try :
        yield redis_conn
    finally:
        await redis_conn.close()
   