import fasttext
import numpy as np
from datetime import datetime


model = fasttext.load_model("university_model_v2.bin")

UNIVERSITY_KEYWORDS = {
    "university", "faculty", "exam", "student", "class", "degree", "registration",
    "professor", "schedule", "grade", "lecture"
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

    if prob > threshold or sim > 0.4:

        print(f"{label} (prob={prob:.2f}, sim={sim:.2f})")
    else:
        print(f"→ Route to General Model (prob={prob:.2f}, sim={sim:.2f})")
        log_uncertain_case(user_input, prob, sim, label)


def log_uncertain_case(text, prob, sim, label):
    with open("uncertain_cases.log", "a") as f:
        f.write(f"[{datetime.now()}] ({label}, p={prob:.2f}, s={sim:.2f}) -> {text}\n")


if __name__ == "__main__":
    print("Router started (Ctrl+C to exit)")
    while True:
        try:
            text = input("User: ").strip()
            if not text:
                continue
            if text.lower() == "exit":
                break
            route_input(text)
        except KeyboardInterrupt:
            break
