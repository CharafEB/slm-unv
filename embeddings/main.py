from fastapi import FastAPI, WebSocket, WebSocketDisconnect ,Depends
from fastapi.responses import JSONResponse
from classifier import mcp_req
import redis.asyncio as redis 
from connection import redis_connection
app = FastAPI()

@app.websocket("/ws/{user_id}")
async def websocket_endpoint(websocket: WebSocket, user_id: str , redis_conn: redis.Redis =Depends(redis_connection)):
    await websocket.accept()
    
    pubsub = redis_conn.pubsub()
    await pubsub.subscribe(f"user_{user_id}_responses")
    try:
        while True:
            message = await pubsub.get_message(
                ignore_subscribe_messages=True, 
                timeout=1.0 
            )

            if message:
                await websocket.send_text(message['data'])
    except WebSocketDisconnect:
        await pubsub.unsubscribe(f"user_{user_id}_responses")


@app.post("/message/{user_id}")
async def send_message(user_id: str, message: str ):
    try:
        return await mcp_req(f"user_{user_id}_responses" , message)
    except Exception as e:
        return JSONResponse(
            status_code=500,
            content={"status": "error", "message": str(e)}
        )