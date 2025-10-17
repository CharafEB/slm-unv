import fasttext

#tain the model
model = fasttext.train_supervised('data.txt', epoch=100, lr=0.25, wordNgrams=3, dim=200,minn=3,maxn=6)

#save the model
model.save_model("university_model_v2.bin")
