import ollama

class UniversityAssistant:
    def __init__(self):
        self.model = "qwen2.5:1.5b-instruct"
        
        self.tools = [{
            'type': 'function',
            'function': {
                'name': 'search_database',
                'description': 'Search university database',
                'parameters': {
                    'type': 'object',
                    'properties': {
                        'query': {
                            'type': 'string',
                            'description': 'Search query'
                        }
                    },
                    'required': ['query']
                }
            }
        }]
    
    def needs_search(self, message):
        """
        قرار سريع: هل نحتاج البحث في قاعدة البيانات؟
        """
        msg = message.lower().strip()
        
        # كلمات مفتاحية تدل على سؤال جامعي
        search_keywords = [
            # عربي
            'تسجيل', 'تسجيلات', 'كتاب', 'كتب', 'مقال', 'مقالات',
            'خبر', 'أخبار', 'جدول', 'جداول', 'امتحان', 'امتحانات',
            'قسم', 'أقسام', 'أستاذ', 'أساتذة', 'دكتور', 'كلية',
            'موعد', 'مواعيد', 'متى', 'أين', 'هل يوجد',
            # إنجليزي
            'registration', 'book', 'article', 'news', 'schedule',
            'exam', 'department', 'professor', 'faculty', 'when', 'where'
        ]
        
        # إذا وجدنا أي كلمة مفتاحية → نحتاج بحث
        return any(keyword in msg for keyword in search_keywords)
    
    def chat(self, message, history=None):
        """
        المحادثة الرئيسية - مع قرار ذكي
        """
        if history is None:
            history = []
        
        # أضف رسالة المستخدم
        history.append({'role': 'user', 'content': message})
        
        # 🎯 القرار الحاسم: هل نحتاج الأداة؟
        if self.needs_search(message):
            # ✅ سؤال يحتاج بحث → استخدم tools
            print("🔍 [استخدام أداة البحث]")
            response = ollama.chat(
                model=self.model,
                messages=history,
                tools=self.tools  # ← الأداة موجودة
            )
        else:
            # ⚡ سؤال روتيني → بدون tools (سريع!)
            print("💬 [رد مباشر]")
            response = ollama.chat(
                model=self.model,
                messages=history
                # ← لاحظ: بدون tools!
            )
        
        # أضف الرد للتاريخ
        assistant_msg = response['message']
        history.append(assistant_msg)
        
        return assistant_msg.get('content', ''), history



def main():
    assistant = UniversityAssistant()
    history = []
    
    print("🎓 مساعد جامعة عبد الحفيظ بوالصوف")
    print("="*50)
    
    while True:
        user_input = input("\n👤 أنت: ").strip()
        
        if user_input.lower() in ['خروج', 'exit', 'quit']:
            print("👋 وداعاً!")
            break
        
        if not user_input:
            continue
        
        # احصل على الرد
        response, history = assistant.chat(user_input, history)
        print(f"🤖 المساعد: {response}")


if __name__ == "__main__":
    main()