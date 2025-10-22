import fasttext
from datetime import datetime
import redis
import os
from dotenv import load_dotenv
import json 

model = fasttext.load_model("university_model_v2.bin")

UNIVERSITY_KEYWORDS = {
    "university", "faculty", "exam", "student", "class", "degree", "registration",
    "professor", "schedule", "grade", "lecture" , "UNV"
}

def RedisConnection():

    load_dotenv("../.env")

    r = redis.Redis(
    host=os.getenv("REDISADD_PY"),
    port=int(os.getenv("REDISPORT")),
    username=os.getenv("REDISUSERNAME"),
    password=os.getenv("REDISPASSWORD"),
    decode_responses=True
)

    channel = "news_channel"
    print("Connected to Redis...")
    return r , channel;


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


if __name__ == "__main__":
    print("Router started (Ctrl+C to exit)")
    con, channel = RedisConnection()
    try:
        while True:
            text = input("You: ").strip()
            if not text:
                continue
            if text.lower() == "exit":
                break
            #print(route_input(text))
            con.publish(channel, json.dumps({"label":route_input(text) , "input":text} , indent=2))
    except KeyboardInterrupt:
        print("\n Goodby")
    finally:
        con.close()
