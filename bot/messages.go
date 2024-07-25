package bot

const (
	StartMessage = `Привет! Я твой финансовый помощник. Вот команды, которые я могу выполнить:
- /addexpense - Добавить расход
- /getexpenses - Показать все расходы
- /addcategory - Добавить новую категорию
- /setlimit - Установить лимит для категории
- /getcategories - Показать все категории
- /deletelastexpense - Удалить последний расход
- /help - Показать справку
`

	HelpMessage = `Я могу помочь тебе управлять личными финансами. Вот список доступных команд:
- /addexpense - Добавить новый расход
- /getexpenses - Показать все расходы
- /addcategory - Добавить новую категорию
- /setlimit - Установить лимит для категории
- /getcategories - Показать все категории
- /deletelastexpense - Удалить последний расход
- /help - Показать это сообщение
`

	// AddExpenseUsage - сообщение с инструкцией по добавлению расхода
	AddExpenseUsage = `Использование: /addexpense <amount> <note>
Пример: /addexpense 100.50 Обед`

	// InvalidAmount - сообщение при вводе неверной суммы
	InvalidAmount = `Неверная сумма. Пожалуйста, введите число, например 100.50`

	// ExpenseAdded - сообщение об успешном добавлении расхода
	ExpenseAdded = `Расход добавлен!`

	// FailedToAddExpense - сообщение об ошибке при добавлении расхода
	FailedToAddExpense = `Ошибка при добавлении расхода. Попробуйте снова.`

	// NoExpensesFound - сообщение при отсутствии расходов
	NoExpensesFound = `Расходы не найдены`

	// GetExpensesHeader - заголовок для списка расходов
	GetExpensesHeader = `Твои расходы:
`

	// UnknownCommand - сообщение при вводе неизвестной команды
	UnknownCommand = `Неизвестная команда. Введите /help для получения списка доступных команд.`
)
