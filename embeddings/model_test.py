import fasttext

#test the model
model = fasttext.load_model("university_model_v2.bin")

test_cases_with_expected = [
    # --- التمييز بين المؤلف والموضوع (بدون أسماء من بيانات التدريب) ---
    ("find me papers by Dr. Karim", "__label__author_article"),
    ("find me papers about Dr. Karim", "__label__article_search"),
    ("search for articles by Professor Lena", "__label__author_article"),
    ("search for articles about Professor Lena", "__label__article_search"),
    ("papers written by Dr. Nabil", "__label__author_article"),
    ("papers discussing Dr. Nabil's framework", "__label__article_search"),
    ("show me research on vision systems by Dr. Mira", "__label__author_article"),
    ("show me research on Dr. Mira's vision systems", "__label__article_search"),

    # --- صيغ عامة / أسئلة / أفعال غير شائعة ---
    ("retrieve all work from Dr. Tarek", "__label__author_article"),
    ("pull up studies on federated learning", "__label__article_search"),
    ("locate anything published by Professor Sami", "__label__author_article"),
    ("i need research regarding knowledge graphs", "__label__article_search"),
    ("can you show me what Dr. Huda has written?", "__label__author_article"),
    ("what has Professor Nour published in robotics?", "__label__author_article"),

    # --- فئة general (أسئلة تعريفية أو تفاعل عام) ---
    ("hello there", "__label__general"),
    ("how’s everything going?", "__label__general"),
    ("what is federated learning?", "__label__general"),
    ("tell me about yourself", "__label__general"),
    ("are you a bot?", "__label__general"),
    ("thanks for your help!", "__label__general"),
    ("what’s the meaning of intelligence?", "__label__general"),
    ("can you explain attention mechanisms?", "__label__general"),

    # --- فئة university (تمييز بين استعلام إداري vs بحث أكاديمي) ---
    ("when does registration start?", "__label__university"),
    ("how do I register for courses?", "__label__university"),
    ("find me articles about university registration", "__label__article_search"),
    ("tell me about student housing", "__label__university"),
    ("what is the policy on campus housing?", "__label__university"),
    ("research on student housing models", "__label__article_search"),

    # --- حالات صعبة (Hard negatives / تمييز دقيق) ---
    ("find me a paper by Sami about robotics", "__label__author_article"),
    ("find me a paper about Sami's robotics work", "__label__article_search"),
    ("search for Dr. Lina's latest publication", "__label__author_article"),
    ("search for publications discussing Dr. Lina", "__label__article_search"),
    ("papers on edge AI by Dr. Faisal", "__label__author_article"),
    ("papers on Dr. Faisal and edge AI", "__label__article_search"),
]

for sentence, expected in test_cases_with_expected:
    label, prob = model.predict(sentence)
    predicted = label[0]
    status = "✅" if predicted == expected else "❌"
    print(f"{status} Input: {sentence}")
    print(f"    Expected: {expected} | Predicted: {predicted} | Confidence: {prob[0]:.3f}\n")