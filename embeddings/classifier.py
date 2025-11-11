import fasttext
from datetime import datetime
import json 
from connection import redis_connection
import redis
from fastapi import Depends
model = fasttext.load_model("university_model_v2.bin")

UNIVERSITY_KEYWORDS = {
    "university", "faculty", "exam", "student", "class", "degree", "registration",
    "professor", "schedule", "grade", "lecture" , "UNV"
}



def fallback_similarity(user_input):
   
    tokens = set(user_input.lower().split())
    overlap = len(tokens & UNIVERSITY_KEYWORDS)
    return overlap / max(len(tokens), 1)

def route_input(user_input):
    label, prob = model.predict(user_input)
    label, prob = label[0], prob[0]

    # Adaptive thresholding
    threshold = 0.75 if len(user_input.split()) > 3 else 0.85

    # fallback similarity check
    sim = fallback_similarity(user_input)

    #Check if the probability
    if prob > threshold or sim > 0.4:
        return label;
    else:
        #Save the user_input so that we can check it latter 
        log_uncertain_case(user_input, prob, sim, label)
        return "__label__general";


def log_uncertain_case(text, prob, sim, label):
    with open("uncertain_cases.log", "a") as f:
        f.write(f"[{datetime.now()}] ({label}, p={prob:.2f}, s={sim:.2f}) -> {text}\n")

async def mcp_req(channel :str , text : str ,redis_conn: redis.Redis =Depends(redis_connection)): 
    try:
        redis_conn.publish("new_channel", json.dumps({"label":route_input(text) , "input":text ,"user_channel" : channel} , indent=2))
        return {"status": "success", "message": "Message sent"}
    except Exception as e :
        return {"status": "Bad", "message": str(e)}
    finally:
        redis_conn.close()