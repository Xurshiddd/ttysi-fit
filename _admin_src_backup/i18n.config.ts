export default defineI18nConfig(() => ({
  legacy: false,
  fallbackLocale: 'uz',
  messages: {
    uz: {
      app: { title: 'TTYSI_FIT Admin' },
      nav: {
        dashboard: 'Bosh sahifa', hemis: 'HEMIS sinxron', faculties: 'Fakultetlar',
        departments: 'Kafedralar', groups: 'Guruhlar', users: 'Foydalanuvchilar'
      },
      auth: {
        login: 'Kirish', email: 'Email', password: 'Parol', logout: 'Chiqish',
        signIn: 'Tizimga kirish', error: 'Login yoki parol noto‘g‘ri'
      },
      common: {
        search: 'Qidirish', role: 'Rol', active: 'Faol', inactive: 'Nofaol',
        all: 'Barchasi', total: 'Jami', name: 'Nomi', code: 'Kodi', faculty: 'Fakultet',
        group: 'Guruh', course: 'Kurs', position: 'Lavozim', loading: 'Yuklanmoqda...',
        empty: 'Ma’lumot yo‘q', actions: 'Amallar', refresh: 'Yangilash'
      },
      roles: { student: 'Talaba', employee: 'Xodim', teacher: 'O‘qituvchi', admin: 'Admin' },
      hemis: {
        title: 'HEMIS sinxronizatsiya',
        desc: 'Tashkiliy tuzilma va foydalanuvchilarni HEMIS’dan yuklash',
        structures: 'Strukturalar', groups: 'Guruhlar', students: 'Talabalar',
        employees: 'Xodimlar', sync: 'Sinxronlash', syncing: 'Sinxronlanmoqda...',
        success: 'Muvaffaqiyatli', created: 'Yangi', updated: 'Yangilangan'
      }
    },
    ru: {
      app: { title: 'TTYSI_FIT Админ' },
      nav: {
        dashboard: 'Главная', hemis: 'HEMIS синхрон', faculties: 'Факультеты',
        departments: 'Кафедры', groups: 'Группы', users: 'Пользователи'
      },
      auth: {
        login: 'Вход', email: 'Email', password: 'Пароль', logout: 'Выход',
        signIn: 'Войти в систему', error: 'Неверный логин или пароль'
      },
      common: {
        search: 'Поиск', role: 'Роль', active: 'Активен', inactive: 'Неактивен',
        all: 'Все', total: 'Всего', name: 'Название', code: 'Код', faculty: 'Факультет',
        group: 'Группа', course: 'Курс', position: 'Должность', loading: 'Загрузка...',
        empty: 'Нет данных', actions: 'Действия', refresh: 'Обновить'
      },
      roles: { student: 'Студент', employee: 'Сотрудник', teacher: 'Преподаватель', admin: 'Админ' },
      hemis: {
        title: 'Синхронизация HEMIS',
        desc: 'Загрузка структуры и пользователей из HEMIS',
        structures: 'Структуры', groups: 'Группы', students: 'Студенты',
        employees: 'Сотрудники', sync: 'Синхронизировать', syncing: 'Синхронизация...',
        success: 'Успешно', created: 'Новые', updated: 'Обновлено'
      }
    },
    en: {
      app: { title: 'TTYSI_FIT Admin' },
      nav: {
        dashboard: 'Dashboard', hemis: 'HEMIS sync', faculties: 'Faculties',
        departments: 'Departments', groups: 'Groups', users: 'Users'
      },
      auth: {
        login: 'Login', email: 'Email', password: 'Password', logout: 'Logout',
        signIn: 'Sign in', error: 'Invalid login or password'
      },
      common: {
        search: 'Search', role: 'Role', active: 'Active', inactive: 'Inactive',
        all: 'All', total: 'Total', name: 'Name', code: 'Code', faculty: 'Faculty',
        group: 'Group', course: 'Course', position: 'Position', loading: 'Loading...',
        empty: 'No data', actions: 'Actions', refresh: 'Refresh'
      },
      roles: { student: 'Student', employee: 'Employee', teacher: 'Teacher', admin: 'Admin' },
      hemis: {
        title: 'HEMIS synchronization',
        desc: 'Import organizational structure and users from HEMIS',
        structures: 'Structures', groups: 'Groups', students: 'Students',
        employees: 'Employees', sync: 'Sync', syncing: 'Syncing...',
        success: 'Success', created: 'Created', updated: 'Updated'
      }
    }
  }
}))
