-- Spanish OS Database Schema
-- FSRS-based spaced repetition with gamification

-- Users table (single user for now, but extensible)
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    total_xp INTEGER DEFAULT 0,
    current_streak INTEGER DEFAULT 0,
    longest_streak INTEGER DEFAULT 0,
    last_active_date DATE
);

-- Islands (vocabulary categories/scripts)
CREATE TABLE IF NOT EXISTS islands (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    icon TEXT,
    unlock_xp INTEGER DEFAULT 0,
    sort_order INTEGER DEFAULT 0
);

-- Cards (vocabulary items)
CREATE TABLE IF NOT EXISTS cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    island_id INTEGER REFERENCES islands(id),
    term TEXT NOT NULL,
    translation TEXT NOT NULL,
    example_sentence TEXT,
    notes TEXT,
    audio_url TEXT,
    frequency_rank INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Bridges (polyglot connections: Hindi/Dutch/English)
CREATE TABLE IF NOT EXISTS bridges (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_id INTEGER NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    bridge_type TEXT NOT NULL CHECK(bridge_type IN ('hindi_phonetic', 'dutch_syntax', 'english_cognate')),
    bridge_content TEXT NOT NULL,
    explanation TEXT
);

-- Card Progress (FSRS scheduling data per user per card)
CREATE TABLE IF NOT EXISTS card_progress (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    card_id INTEGER NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    stability REAL DEFAULT 0,
    difficulty REAL DEFAULT 0,
    elapsed_days INTEGER DEFAULT 0,
    scheduled_days INTEGER DEFAULT 0,
    reps INTEGER DEFAULT 0,
    lapses INTEGER DEFAULT 0,
    state TEXT DEFAULT 'new' CHECK(state IN ('new', 'learning', 'review', 'relearning')),
    due DATETIME,
    last_review DATETIME,
    UNIQUE(user_id, card_id)
);

-- Review Logs (history for FSRS optimization and analytics)
CREATE TABLE IF NOT EXISTS review_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    card_id INTEGER NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK(rating IN (1, 2, 3, 4)),
    elapsed_days INTEGER,
    scheduled_days INTEGER,
    reviewed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    review_duration_ms INTEGER
);

-- Daily Logs (for heat map and daily stats)
CREATE TABLE IF NOT EXISTS daily_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    date DATE NOT NULL,
    xp_earned INTEGER DEFAULT 0,
    cards_reviewed INTEGER DEFAULT 0,
    cards_correct INTEGER DEFAULT 0,
    minutes_active INTEGER DEFAULT 0,
    new_cards_added INTEGER DEFAULT 0,
    UNIQUE(user_id, date)
);

-- Achievements (gamification badges)
CREATE TABLE IF NOT EXISTS achievements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    icon TEXT,
    xp_reward INTEGER DEFAULT 0,
    condition_type TEXT,
    condition_value INTEGER
);

-- User Achievements (earned badges)
CREATE TABLE IF NOT EXISTS user_achievements (
    user_id INTEGER NOT NULL REFERENCES users(id),
    achievement_id INTEGER NOT NULL REFERENCES achievements(id),
    earned_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(user_id, achievement_id)
);

-- User Settings (JSON storage for flexibility)
CREATE TABLE IF NOT EXISTS user_settings (
    user_id INTEGER PRIMARY KEY REFERENCES users(id),
    settings TEXT NOT NULL DEFAULT '{}'
);

-- Curriculum Journey (track user's 180-day journey)
CREATE TABLE IF NOT EXISTS curriculum_journey (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    start_date DATE NOT NULL,
    is_active INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

-- Lesson Sessions (track daily lesson completions)
CREATE TABLE IF NOT EXISTS lesson_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    session_date DATE NOT NULL,
    day_number INTEGER NOT NULL,
    phase_id INTEGER NOT NULL,
    cards_reviewed INTEGER DEFAULT 0,
    cards_correct INTEGER DEFAULT 0,
    new_cards_learned INTEGER DEFAULT 0,
    xp_earned INTEGER DEFAULT 0,
    completed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, session_date)
);

-- Questions (AI-generated question types for cards)
CREATE TABLE IF NOT EXISTS questions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_id INTEGER NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    question_type TEXT NOT NULL CHECK(question_type IN ('mcq', 'fill_blank', 'sentence_build')),
    question_data TEXT NOT NULL, -- JSON with type-specific data
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Grammar Rules (AI-generated grammar explanations)
CREATE TABLE IF NOT EXISTS grammar_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rule_key TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    explanation TEXT NOT NULL,
    examples TEXT NOT NULL, -- JSON array of examples
    related_cards TEXT, -- JSON array of card IDs
    difficulty_level INTEGER DEFAULT 1 CHECK(difficulty_level IN (1, 2, 3)),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Card to Grammar Rule mapping
CREATE TABLE IF NOT EXISTS card_grammar (
    card_id INTEGER NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    grammar_rule_id INTEGER NOT NULL REFERENCES grammar_rules(id) ON DELETE CASCADE,
    PRIMARY KEY (card_id, grammar_rule_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_card_progress_due ON card_progress(due);
CREATE INDEX IF NOT EXISTS idx_card_progress_user_state ON card_progress(user_id, state);
CREATE INDEX IF NOT EXISTS idx_review_logs_user_card ON review_logs(user_id, card_id);
CREATE INDEX IF NOT EXISTS idx_daily_logs_user_date ON daily_logs(user_id, date);
CREATE INDEX IF NOT EXISTS idx_cards_island ON cards(island_id);
CREATE INDEX IF NOT EXISTS idx_cards_frequency ON cards(frequency_rank);
CREATE INDEX IF NOT EXISTS idx_bridges_card ON bridges(card_id);
CREATE INDEX IF NOT EXISTS idx_cards_term ON cards(term);
CREATE INDEX IF NOT EXISTS idx_cards_translation ON cards(translation);
CREATE INDEX IF NOT EXISTS idx_curriculum_journey_user ON curriculum_journey(user_id);
CREATE INDEX IF NOT EXISTS idx_lesson_sessions_user_date ON lesson_sessions(user_id, session_date);
CREATE INDEX IF NOT EXISTS idx_questions_card ON questions(card_id);
CREATE INDEX IF NOT EXISTS idx_questions_type ON questions(question_type);
CREATE INDEX IF NOT EXISTS idx_grammar_difficulty ON grammar_rules(difficulty_level);
CREATE INDEX IF NOT EXISTS idx_card_grammar_card ON card_grammar(card_id);

-- Seed Data: Default user
INSERT OR IGNORE INTO users (id, username) VALUES (1, 'sangam');

-- Seed Data: Islands (vocabulary categories)
INSERT OR IGNORE INTO islands (id, name, description, icon, unlock_xp, sort_order) VALUES
    (1, 'Core Essentials', 'Top 100 most frequent words', '1', 0, 1),
    (2, 'Common Words', 'Words 101-250 by frequency', '2', 500, 2),
    (3, 'Expanding Vocabulary', 'Words 251-500 by frequency', '3', 1500, 3),
    (4, 'Advanced Vocabulary', 'Words 501-1000 by frequency', '4', 3000, 4),
    (5, 'Core Verbs', 'Essential action words', '5', 500, 5),
    (6, 'Career & Work', 'Professional vocabulary', '6', 1000, 6),
    (7, 'Social & Daily Life', 'Everyday conversations', '7', 1000, 7),
    (8, 'Past Tense', 'Narrating the past', '8', 2000, 8),
    (9, 'Subjunctive', 'Wishes, hopes, and uncertainty', '9', 5000, 9);

-- Seed Data: Sample achievements
INSERT OR IGNORE INTO achievements (id, name, description, icon, xp_reward, condition_type, condition_value) VALUES
    (1, 'First Steps', 'Review your first card', 'sparkles', 10, 'cards_reviewed', 1),
    (2, 'Getting Started', 'Review 10 cards', 'seedling', 25, 'cards_reviewed', 10),
    (3, 'Dedicated Learner', 'Review 100 cards', 'books', 100, 'cards_reviewed', 100),
    (4, 'Century Club', 'Review 500 cards', 'hundred', 250, 'cards_reviewed', 500),
    (5, 'Vocabulary Master', 'Review 1000 cards', 'crown', 500, 'cards_reviewed', 1000),
    (6, 'Week Warrior', '7-day streak', 'fire', 50, 'streak', 7),
    (7, 'Fortnight Fighter', '14-day streak', 'zap', 100, 'streak', 14),
    (8, 'Month Master', '30-day streak', 'trophy', 200, 'streak', 30),
    (9, 'Streak Legend', '100-day streak', 'star', 1000, 'streak', 100),
    (10, 'Word Collector', 'Learn 50 words', 'gem', 75, 'words_learned', 50),
    (11, 'Polyglot Path', 'Learn 200 words', 'rocket', 200, 'words_learned', 200),
    (12, 'Lexicon Lord', 'Learn 500 words', 'brain', 500, 'words_learned', 500);

-- Seed Data: Spanish Vocabulary
-- Island 1: Core Essentials (Top 100 most frequent words)
INSERT OR IGNORE INTO cards (island_id, term, translation, frequency_rank, example_sentence) VALUES
    (1, 'el', 'the (masculine)', 1, 'El libro está en la mesa.'),
    (1, 'de', 'of, from', 2, 'Soy de México.'),
    (1, 'que', 'that, which', 3, 'Creo que sí.'),
    (1, 'y', 'and', 4, 'Tú y yo somos amigos.'),
    (1, 'a', 'to, at', 5, 'Voy a la escuela.'),
    (1, 'en', 'in, on', 6, 'Estoy en casa.'),
    (1, 'un', 'a, an (masculine)', 7, 'Tengo un perro.'),
    (1, 'ser', 'to be (permanent)', 8, 'Soy estudiante.'),
    (1, 'se', 'himself/herself/itself', 9, 'Se llama María.'),
    (1, 'no', 'no, not', 10, 'No entiendo.'),
    (1, 'haber', 'to have (auxiliary)', 11, 'He comido ya.'),
    (1, 'por', 'for, by, through', 12, 'Gracias por todo.'),
    (1, 'con', 'with', 13, 'Café con leche.'),
    (1, 'su', 'his, her, their, your (formal)', 14, 'Su casa es grande.'),
    (1, 'para', 'for, in order to', 15, 'Es para ti.'),
    (1, 'como', 'like, as, how', 16, '¿Cómo estás?'),
    (1, 'estar', 'to be (temporary)', 17, 'Estoy cansado.'),
    (1, 'tener', 'to have', 18, 'Tengo hambre.'),
    (1, 'le', 'him, her, you (indirect)', 19, 'Le di el libro.'),
    (1, 'lo', 'him, it (direct)', 20, 'Lo veo.'),
    (1, 'todo', 'all, everything', 21, 'Todo está bien.'),
    (1, 'pero', 'but', 22, 'Quiero ir, pero no puedo.'),
    (1, 'más', 'more, plus', 23, 'Quiero más agua.'),
    (1, 'hacer', 'to do, to make', 24, '¿Qué haces?'),
    (1, 'o', 'or', 25, '¿Café o té?'),
    (1, 'poder', 'to be able to, can', 26, 'Puedo ayudarte.'),
    (1, 'decir', 'to say, to tell', 27, '¿Qué dices?'),
    (1, 'este', 'this (masculine)', 28, 'Este libro es mío.'),
    (1, 'ir', 'to go', 29, 'Voy al cine.'),
    (1, 'otro', 'other, another', 30, 'Dame otro café.'),
    (1, 'ese', 'that (masculine)', 31, 'Ese carro es rojo.'),
    (1, 'la', 'the (feminine)', 32, 'La casa es bonita.'),
    (1, 'si', 'if, yes', 33, 'Si quieres, vamos.'),
    (1, 'me', 'me', 34, 'Me gusta el chocolate.'),
    (1, 'ya', 'already, now', 35, 'Ya terminé.'),
    (1, 'ver', 'to see', 36, 'Quiero ver la película.'),
    (1, 'porque', 'because', 37, 'No voy porque estoy enfermo.'),
    (1, 'dar', 'to give', 38, 'Dame tu número.'),
    (1, 'cuando', 'when', 39, '¿Cuándo vienes?'),
    (1, 'él', 'he, him', 40, 'Él es mi hermano.'),
    (1, 'muy', 'very', 41, 'Es muy interesante.'),
    (1, 'sin', 'without', 42, 'Café sin azúcar.'),
    (1, 'vez', 'time (occasion)', 43, 'Una vez más.'),
    (1, 'mucho', 'much, a lot', 44, 'Te quiero mucho.'),
    (1, 'saber', 'to know', 45, 'No sé la respuesta.'),
    (1, 'qué', 'what', 46, '¿Qué quieres?'),
    (1, 'sobre', 'about, on, over', 47, 'Hablamos sobre el trabajo.'),
    (1, 'mi', 'my', 48, 'Mi familia es grande.'),
    (1, 'alguno', 'some, any', 49, '¿Tienes alguna pregunta?'),
    (1, 'mismo', 'same, self', 50, 'Es lo mismo.'),
    (1, 'yo', 'I', 51, 'Yo soy estudiante.'),
    (1, 'también', 'also, too', 52, 'Yo también quiero ir.'),
    (1, 'hasta', 'until, even', 53, 'Hasta mañana.'),
    (1, 'año', 'year', 54, 'Feliz año nuevo.'),
    (1, 'dos', 'two', 55, 'Tengo dos hermanos.'),
    (1, 'querer', 'to want, to love', 56, 'Te quiero.'),
    (1, 'entre', 'between, among', 57, 'Entre tú y yo.'),
    (1, 'así', 'like this, thus', 58, 'Hazlo así.'),
    (1, 'primero', 'first', 59, 'Primero, desayuno.'),
    (1, 'desde', 'from, since', 60, 'Desde ayer.'),
    (1, 'grande', 'big, large', 61, 'Una casa grande.'),
    (1, 'eso', 'that (neuter)', 62, '¿Qué es eso?'),
    (1, 'ni', 'neither, nor', 63, 'Ni tú ni yo.'),
    (1, 'nos', 'us', 64, 'Nos vemos mañana.'),
    (1, 'llegar', 'to arrive', 65, 'Voy a llegar tarde.'),
    (1, 'pasar', 'to pass, to happen', 66, '¿Qué pasó?'),
    (1, 'tiempo', 'time, weather', 67, 'No tengo tiempo.'),
    (1, 'ella', 'she, her', 68, 'Ella es mi amiga.'),
    (1, 'sí', 'yes', 69, 'Sí, claro.'),
    (1, 'día', 'day', 70, 'Buenos días.'),
    (1, 'uno', 'one', 71, 'Dame uno.'),
    (1, 'bien', 'well, good', 72, 'Estoy bien.'),
    (1, 'poco', 'little, few', 73, 'Un poco de agua.'),
    (1, 'deber', 'must, to owe', 74, 'Debo estudiar.'),
    (1, 'entonces', 'then', 75, 'Entonces, vamos.'),
    (1, 'poner', 'to put', 76, 'Pon la mesa.'),
    (1, 'cosa', 'thing', 77, '¿Qué cosa?'),
    (1, 'tanto', 'so much', 78, 'No es para tanto.'),
    (1, 'hombre', 'man', 79, 'El hombre camina.'),
    (1, 'parecer', 'to seem, to appear', 80, 'Parece difícil.'),
    (1, 'nuestro', 'our', 81, 'Nuestra casa.'),
    (1, 'tan', 'so, as', 82, 'Es tan fácil.'),
    (1, 'donde', 'where', 83, '¿Dónde estás?'),
    (1, 'ahora', 'now', 84, 'Ahora mismo.'),
    (1, 'parte', 'part', 85, 'En alguna parte.'),
    (1, 'después', 'after, later', 86, 'Después de comer.'),
    (1, 'vida', 'life', 87, 'La vida es bella.'),
    (1, 'quedar', 'to stay, to remain', 88, 'Me quedo aquí.'),
    (1, 'siempre', 'always', 89, 'Siempre te amaré.'),
    (1, 'creer', 'to believe', 90, 'Creo que sí.'),
    (1, 'hablar', 'to speak, to talk', 91, 'Hablo español.'),
    (1, 'llevar', 'to carry, to wear', 92, 'Llevo una camisa azul.'),
    (1, 'dejar', 'to leave, to let', 93, 'Déjame en paz.'),
    (1, 'nada', 'nothing', 94, 'No pasa nada.'),
    (1, 'cada', 'each, every', 95, 'Cada día aprendo algo.'),
    (1, 'seguir', 'to follow, to continue', 96, 'Sigue adelante.'),
    (1, 'menos', 'less, minus', 97, 'Más o menos.'),
    (1, 'nuevo', 'new', 98, 'Tengo un carro nuevo.'),
    (1, 'encontrar', 'to find', 99, 'No puedo encontrarlo.'),
    (1, 'tres', 'three', 100, 'Tengo tres gatos.');

-- Island 2: Common Words (101-250 frequency)
INSERT OR IGNORE INTO cards (island_id, term, translation, frequency_rank, example_sentence) VALUES
    (2, 'bueno', 'good', 101, '¡Qué bueno!'),
    (2, 'venir', 'to come', 102, 'Ven aquí.'),
    (2, 'pensar', 'to think', 103, 'Pienso en ti.'),
    (2, 'salir', 'to go out, to leave', 104, 'Salgo a las ocho.'),
    (2, 'volver', 'to return', 105, 'Vuelvo pronto.'),
    (2, 'tomar', 'to take, to drink', 106, 'Tomo café.'),
    (2, 'conocer', 'to know (person/place)', 107, '¿Conoces Madrid?'),
    (2, 'sentir', 'to feel', 108, 'Lo siento mucho.'),
    (2, 'tratar', 'to try, to treat', 109, 'Trato de entender.'),
    (2, 'mirar', 'to look at', 110, 'Mírame.'),
    (2, 'contar', 'to count, to tell', 111, 'Cuéntame todo.'),
    (2, 'empezar', 'to begin, to start', 112, 'Empezamos mañana.'),
    (2, 'esperar', 'to wait, to hope', 113, 'Espero que sí.'),
    (2, 'buscar', 'to look for', 114, 'Busco mis llaves.'),
    (2, 'entrar', 'to enter', 115, 'Entra por favor.'),
    (2, 'trabajar', 'to work', 116, 'Trabajo mucho.'),
    (2, 'escribir', 'to write', 117, 'Escribo una carta.'),
    (2, 'perder', 'to lose', 118, 'No quiero perder.'),
    (2, 'producir', 'to produce', 119, 'Producen vino.'),
    (2, 'ocurrir', 'to occur, to happen', 120, '¿Qué ocurrió?'),
    (2, 'entender', 'to understand', 121, 'No entiendo.'),
    (2, 'pedir', 'to ask for, to order', 122, 'Pido la cuenta.'),
    (2, 'recibir', 'to receive', 123, 'Recibí tu mensaje.'),
    (2, 'recordar', 'to remember', 124, '¿Recuerdas?'),
    (2, 'terminar', 'to finish', 125, 'Terminé el trabajo.'),
    (2, 'permitir', 'to allow', 126, 'No lo permito.'),
    (2, 'aparecer', 'to appear', 127, 'Apareció de repente.'),
    (2, 'conseguir', 'to get, to achieve', 128, 'Conseguí el trabajo.'),
    (2, 'comenzar', 'to begin', 129, 'Comenzamos ahora.'),
    (2, 'servir', 'to serve', 130, '¿En qué puedo servirle?'),
    (2, 'casa', 'house', 131, 'Mi casa es tu casa.'),
    (2, 'mundo', 'world', 132, 'El mundo es pequeño.'),
    (2, 'país', 'country', 133, 'Mi país es hermoso.'),
    (2, 'lugar', 'place', 134, 'Este es un buen lugar.'),
    (2, 'persona', 'person', 135, 'Es una buena persona.'),
    (2, 'momento', 'moment', 136, 'Un momento, por favor.'),
    (2, 'forma', 'form, way', 137, 'De esta forma.'),
    (2, 'punto', 'point', 138, 'Buen punto.'),
    (2, 'gobierno', 'government', 139, 'El gobierno decide.'),
    (2, 'trabajo', 'work, job', 140, 'Tengo mucho trabajo.'),
    (2, 'hecho', 'fact, done', 141, 'De hecho, es verdad.'),
    (2, 'ejemplo', 'example', 142, 'Por ejemplo.'),
    (2, 'lado', 'side', 143, 'Al otro lado.'),
    (2, 'niño', 'child, boy', 144, 'El niño juega.'),
    (2, 'manera', 'manner, way', 145, 'De alguna manera.'),
    (2, 'palabra', 'word', 146, 'Una palabra más.'),
    (2, 'mano', 'hand', 147, 'Dame la mano.'),
    (2, 'mujer', 'woman, wife', 148, 'La mujer trabaja.'),
    (2, 'agua', 'water', 149, 'Quiero agua.'),
    (2, 'razón', 'reason', 150, 'Tienes razón.'),
    (2, 'problema', 'problem', 151, 'No hay problema.'),
    (2, 'grupo', 'group', 152, 'Un grupo de amigos.'),
    (2, 'cuatro', 'four', 153, 'Cuatro estaciones.'),
    (2, 'cinco', 'five', 154, 'Cinco minutos.'),
    (2, 'dentro', 'inside', 155, 'Dentro de la casa.'),
    (2, 'bajo', 'under, low', 156, 'Bajo la mesa.'),
    (2, 'alto', 'tall, high', 157, 'Muy alto.'),
    (2, 'mientras', 'while', 158, 'Mientras espero.'),
    (2, 'aunque', 'although', 159, 'Aunque llueva.'),
    (2, 'casi', 'almost', 160, 'Casi terminé.'),
    (2, 'mejor', 'better, best', 161, 'Es lo mejor.'),
    (2, 'según', 'according to', 162, 'Según yo.'),
    (2, 'solo', 'alone, only', 163, 'Estoy solo.'),
    (2, 'aquí', 'here', 164, 'Ven aquí.'),
    (2, 'cierto', 'certain, true', 165, 'Es cierto.'),
    (2, 'claro', 'clear, of course', 166, '¡Claro que sí!'),
    (2, 'junto', 'together', 167, 'Juntos somos fuertes.'),
    (2, 'único', 'unique, only', 168, 'Eres único.'),
    (2, 'todavía', 'still, yet', 169, 'Todavía no.'),
    (2, 'durante', 'during', 170, 'Durante la noche.'),
    (2, 'madre', 'mother', 171, 'Mi madre cocina.'),
    (2, 'padre', 'father', 172, 'Mi padre trabaja.'),
    (2, 'hermano', 'brother', 173, 'Mi hermano mayor.'),
    (2, 'hermana', 'sister', 174, 'Mi hermana menor.'),
    (2, 'hijo', 'son', 175, 'Tengo un hijo.'),
    (2, 'hija', 'daughter', 176, 'Mi hija estudia.'),
    (2, 'amigo', 'friend', 177, 'Es mi mejor amigo.'),
    (2, 'familia', 'family', 178, 'Amo a mi familia.'),
    (2, 'ciudad', 'city', 179, 'La ciudad es grande.'),
    (2, 'calle', 'street', 180, 'Camino por la calle.'),
    (2, 'libro', 'book', 181, 'Leo un libro.'),
    (2, 'escuela', 'school', 182, 'Voy a la escuela.'),
    (2, 'dinero', 'money', 183, 'No tengo dinero.'),
    (2, 'comida', 'food', 184, 'La comida está rica.'),
    (2, 'carro', 'car', 185, 'Tengo un carro nuevo.'),
    (2, 'noche', 'night', 186, 'Buenas noches.'),
    (2, 'mañana', 'morning, tomorrow', 187, 'Hasta mañana.'),
    (2, 'tarde', 'afternoon, late', 188, 'Buenas tardes.'),
    (2, 'semana', 'week', 189, 'La próxima semana.'),
    (2, 'mes', 'month', 190, 'El mes pasado.'),
    (2, 'hoy', 'today', 191, 'Hoy es lunes.'),
    (2, 'ayer', 'yesterday', 192, 'Ayer fue domingo.'),
    (2, 'negro', 'black', 193, 'El gato negro.'),
    (2, 'blanco', 'white', 194, 'La casa blanca.'),
    (2, 'rojo', 'red', 195, 'El carro rojo.'),
    (2, 'verde', 'green', 196, 'El árbol verde.'),
    (2, 'azul', 'blue', 197, 'El cielo azul.'),
    (2, 'pequeño', 'small', 198, 'Un perro pequeño.'),
    (2, 'largo', 'long', 199, 'Un camino largo.'),
    (2, 'joven', 'young', 200, 'Soy joven todavía.');

-- Island 3: Expanding Vocabulary (251-500 frequency)
INSERT OR IGNORE INTO cards (island_id, term, translation, frequency_rank, example_sentence) VALUES
    (3, 'viejo', 'old', 201, 'El edificio viejo.'),
    (3, 'derecho', 'right, law', 202, 'A la derecha.'),
    (3, 'izquierdo', 'left', 203, 'A la izquierda.'),
    (3, 'último', 'last', 204, 'El último día.'),
    (3, 'cerca', 'near', 205, 'Está cerca de aquí.'),
    (3, 'lejos', 'far', 206, 'Está muy lejos.'),
    (3, 'arriba', 'up, above', 207, 'Mira arriba.'),
    (3, 'abajo', 'down, below', 208, 'Mira abajo.'),
    (3, 'adelante', 'forward', 209, 'Sigue adelante.'),
    (3, 'atrás', 'back, behind', 210, 'Mira atrás.'),
    (3, 'temprano', 'early', 211, 'Es muy temprano.'),
    (3, 'cuenta', 'account, bill', 212, 'La cuenta, por favor.'),
    (3, 'puerta', 'door', 213, 'Abre la puerta.'),
    (3, 'ventana', 'window', 214, 'Cierra la ventana.'),
    (3, 'mesa', 'table', 215, 'Pon la mesa.'),
    (3, 'silla', 'chair', 216, 'Siéntate en la silla.'),
    (3, 'cama', 'bed', 217, 'Me voy a la cama.'),
    (3, 'cocina', 'kitchen', 218, 'Cocino en la cocina.'),
    (3, 'baño', 'bathroom', 219, '¿Dónde está el baño?'),
    (3, 'cuarto', 'room', 220, 'Mi cuarto es grande.'),
    (3, 'tienda', 'store', 221, 'Voy a la tienda.'),
    (3, 'restaurante', 'restaurant', 222, 'Vamos al restaurante.'),
    (3, 'hospital', 'hospital', 223, 'Está en el hospital.'),
    (3, 'banco', 'bank', 224, 'Voy al banco.'),
    (3, 'oficina', 'office', 225, 'Trabajo en una oficina.'),
    (3, 'aeropuerto', 'airport', 226, 'Vamos al aeropuerto.'),
    (3, 'hotel', 'hotel', 227, 'Me quedo en un hotel.'),
    (3, 'playa', 'beach', 228, 'Vamos a la playa.'),
    (3, 'montaña', 'mountain', 229, 'La montaña es alta.'),
    (3, 'río', 'river', 230, 'El río es largo.'),
    (3, 'mar', 'sea', 231, 'El mar está tranquilo.'),
    (3, 'cielo', 'sky', 232, 'El cielo está azul.'),
    (3, 'sol', 'sun', 233, 'El sol brilla.'),
    (3, 'luna', 'moon', 234, 'La luna llena.'),
    (3, 'estrella', 'star', 235, 'Las estrellas brillan.'),
    (3, 'lluvia', 'rain', 236, 'Cae la lluvia.'),
    (3, 'nieve', 'snow', 237, 'Cae nieve.'),
    (3, 'viento', 'wind', 238, 'Hace viento.'),
    (3, 'calor', 'heat, hot', 239, 'Hace calor.'),
    (3, 'frío', 'cold', 240, 'Hace frío.'),
    (3, 'perro', 'dog', 241, 'Tengo un perro.'),
    (3, 'gato', 'cat', 242, 'El gato duerme.'),
    (3, 'pájaro', 'bird', 243, 'El pájaro canta.'),
    (3, 'árbol', 'tree', 244, 'El árbol es alto.'),
    (3, 'flor', 'flower', 245, 'La flor es bonita.'),
    (3, 'pan', 'bread', 246, 'Compro pan.'),
    (3, 'leche', 'milk', 247, 'Bebo leche.'),
    (3, 'carne', 'meat', 248, 'No como carne.'),
    (3, 'pescado', 'fish', 249, 'Me gusta el pescado.'),
    (3, 'pollo', 'chicken', 250, 'Pollo asado.'),
    (3, 'arroz', 'rice', 251, 'Arroz con pollo.'),
    (3, 'fruta', 'fruit', 252, 'Como fruta cada día.'),
    (3, 'manzana', 'apple', 253, 'Una manzana roja.'),
    (3, 'naranja', 'orange', 254, 'Jugo de naranja.'),
    (3, 'plátano', 'banana', 255, 'Me gustan los plátanos.'),
    (3, 'verdura', 'vegetable', 256, 'Como verduras.'),
    (3, 'ensalada', 'salad', 257, 'Una ensalada fresca.'),
    (3, 'sopa', 'soup', 258, 'Sopa caliente.'),
    (3, 'café', 'coffee', 259, 'Un café, por favor.'),
    (3, 'té', 'tea', 260, 'Prefiero té.'),
    (3, 'cerveza', 'beer', 261, 'Una cerveza fría.'),
    (3, 'vino', 'wine', 262, 'Vino tinto.'),
    (3, 'desayuno', 'breakfast', 263, 'Tomo desayuno.'),
    (3, 'almuerzo', 'lunch', 264, '¿Vamos a almorzar?'),
    (3, 'cena', 'dinner', 265, 'La cena está lista.'),
    (3, 'camisa', 'shirt', 266, 'Camisa blanca.'),
    (3, 'pantalón', 'pants', 267, 'Pantalón negro.'),
    (3, 'zapato', 'shoe', 268, 'Zapatos nuevos.'),
    (3, 'vestido', 'dress', 269, 'Un vestido bonito.'),
    (3, 'chaqueta', 'jacket', 270, 'Hace frío, trae tu chaqueta.'),
    (3, 'sombrero', 'hat', 271, 'Lleva sombrero.'),
    (3, 'bolsa', 'bag', 272, 'Una bolsa grande.'),
    (3, 'teléfono', 'telephone', 273, '¿Cuál es tu teléfono?'),
    (3, 'computadora', 'computer', 274, 'Trabajo en la computadora.'),
    (3, 'música', 'music', 275, 'Me gusta la música.'),
    (3, 'película', 'movie', 276, 'Vamos a ver una película.'),
    (3, 'programa', 'program', 277, 'Un programa de televisión.'),
    (3, 'juego', 'game', 278, 'Un juego divertido.'),
    (3, 'fiesta', 'party', 279, 'Una fiesta de cumpleaños.'),
    (3, 'regalo', 'gift', 280, 'Un regalo para ti.'),
    (3, 'feliz', 'happy', 281, 'Estoy muy feliz.'),
    (3, 'triste', 'sad', 282, 'Me siento triste.'),
    (3, 'enojado', 'angry', 283, 'Está enojado.'),
    (3, 'cansado', 'tired', 284, 'Estoy cansado.'),
    (3, 'enfermo', 'sick', 285, 'Estoy enfermo.'),
    (3, 'sano', 'healthy', 286, 'Estoy sano.'),
    (3, 'fuerte', 'strong', 287, 'Es muy fuerte.'),
    (3, 'débil', 'weak', 288, 'Me siento débil.'),
    (3, 'difícil', 'difficult', 289, 'Es muy difícil.'),
    (3, 'fácil', 'easy', 290, 'Es muy fácil.'),
    (3, 'posible', 'possible', 291, 'Es posible.'),
    (3, 'imposible', 'impossible', 292, 'Es imposible.'),
    (3, 'importante', 'important', 293, 'Es muy importante.'),
    (3, 'necesario', 'necessary', 294, 'Es necesario.'),
    (3, 'diferente', 'different', 295, 'Es diferente.'),
    (3, 'igual', 'equal, same', 296, 'Es igual.'),
    (3, 'bonito', 'pretty', 297, 'Qué bonito.'),
    (3, 'feo', 'ugly', 298, 'No es feo.'),
    (3, 'rico', 'rich, delicious', 299, '¡Qué rico!'),
    (3, 'pobre', 'poor', 300, 'El hombre pobre.');

-- Island 5: Core Verbs
INSERT OR IGNORE INTO cards (island_id, term, translation, frequency_rank, example_sentence) VALUES
    (5, 'comer', 'to eat', 301, 'Voy a comer.'),
    (5, 'beber', 'to drink', 302, 'Quiero beber agua.'),
    (5, 'dormir', 'to sleep', 303, 'Necesito dormir.'),
    (5, 'despertar', 'to wake up', 304, 'Me despierto temprano.'),
    (5, 'levantar', 'to lift, to get up', 305, 'Me levanto a las siete.'),
    (5, 'sentar', 'to sit', 306, 'Siéntate aquí.'),
    (5, 'caminar', 'to walk', 307, 'Camino al trabajo.'),
    (5, 'correr', 'to run', 308, 'Corro cada mañana.'),
    (5, 'nadar', 'to swim', 309, '¿Sabes nadar?'),
    (5, 'bailar', 'to dance', 310, 'Me gusta bailar.'),
    (5, 'cantar', 'to sing', 311, 'Canta muy bien.'),
    (5, 'jugar', 'to play', 312, 'Los niños juegan.'),
    (5, 'estudiar', 'to study', 313, 'Estudio español.'),
    (5, 'aprender', 'to learn', 314, 'Aprendo rápido.'),
    (5, 'enseñar', 'to teach', 315, 'Enseño inglés.'),
    (5, 'leer', 'to read', 316, 'Leo un libro.'),
    (5, 'comprar', 'to buy', 317, 'Voy a comprar pan.'),
    (5, 'vender', 'to sell', 318, 'Vende carros.'),
    (5, 'pagar', 'to pay', 319, 'Voy a pagar.'),
    (5, 'abrir', 'to open', 320, 'Abre la puerta.'),
    (5, 'cerrar', 'to close', 321, 'Cierra la ventana.'),
    (5, 'ayudar', 'to help', 322, '¿Puedes ayudarme?'),
    (5, 'necesitar', 'to need', 323, 'Necesito ayuda.'),
    (5, 'usar', 'to use', 324, '¿Cómo se usa?'),
    (5, 'llamar', 'to call', 325, 'Te llamo después.'),
    (5, 'escuchar', 'to listen', 326, 'Escucha la música.'),
    (5, 'oír', 'to hear', 327, 'No oigo nada.'),
    (5, 'preguntar', 'to ask', 328, 'Quiero preguntar algo.'),
    (5, 'responder', 'to answer', 329, 'Responde la pregunta.'),
    (5, 'vivir', 'to live', 330, 'Vivo en Madrid.'),
    (5, 'morir', 'to die', 331, 'No quiero morir.'),
    (5, 'nacer', 'to be born', 332, 'Nací en México.'),
    (5, 'crecer', 'to grow', 333, 'Los niños crecen rápido.'),
    (5, 'cambiar', 'to change', 334, 'Todo cambia.'),
    (5, 'mejorar', 'to improve', 335, 'Quiero mejorar mi español.'),
    (5, 'preparar', 'to prepare', 336, 'Preparo la cena.'),
    (5, 'cocinar', 'to cook', 337, 'Me gusta cocinar.'),
    (5, 'limpiar', 'to clean', 338, 'Limpio la casa.'),
    (5, 'lavar', 'to wash', 339, 'Lavo la ropa.'),
    (5, 'secar', 'to dry', 340, 'Seco los platos.'),
    (5, 'romper', 'to break', 341, 'No lo rompas.'),
    (5, 'arreglar', 'to fix, to arrange', 342, 'Arreglo el carro.'),
    (5, 'intentar', 'to try', 343, 'Voy a intentarlo.'),
    (5, 'lograr', 'to achieve', 344, 'Logré mi meta.'),
    (5, 'ganar', 'to win, to earn', 345, 'Quiero ganar.'),
    (5, 'perder', 'to lose', 346, 'No quiero perder.'),
    (5, 'olvidar', 'to forget', 347, 'No olvides.'),
    (5, 'acordar', 'to agree, to remember', 348, 'Me acuerdo de ti.'),
    (5, 'decidir', 'to decide', 349, 'Tú decides.'),
    (5, 'elegir', 'to choose', 350, 'Elige uno.');

-- Island 7: Social & Daily Life
INSERT OR IGNORE INTO cards (island_id, term, translation, frequency_rank, example_sentence) VALUES
    (7, 'gracias', 'thank you', 351, 'Muchas gracias.'),
    (7, 'por favor', 'please', 352, 'Agua, por favor.'),
    (7, 'de nada', 'you''re welcome', 353, 'De nada.'),
    (7, 'perdón', 'sorry, excuse me', 354, 'Perdón, ¿puedo pasar?'),
    (7, 'disculpa', 'excuse me, sorry', 355, 'Disculpa la molestia.'),
    (7, 'hola', 'hello', 356, '¡Hola! ¿Cómo estás?'),
    (7, 'adiós', 'goodbye', 357, 'Adiós, hasta luego.'),
    (7, 'hasta luego', 'see you later', 358, 'Hasta luego.'),
    (7, 'buenas noches', 'good night', 359, 'Buenas noches.'),
    (7, 'buenos días', 'good morning', 360, 'Buenos días.'),
    (7, 'buenas tardes', 'good afternoon', 361, 'Buenas tardes.'),
    (7, 'bienvenido', 'welcome', 362, '¡Bienvenido a mi casa!'),
    (7, 'felicidades', 'congratulations', 363, '¡Felicidades!'),
    (7, 'feliz cumpleaños', 'happy birthday', 364, '¡Feliz cumpleaños!'),
    (7, 'salud', 'health, bless you', 365, '¡Salud!'),
    (7, 'con permiso', 'excuse me', 366, 'Con permiso.'),
    (7, 'encantado', 'pleased to meet you', 367, 'Encantado de conocerte.'),
    (7, 'mucho gusto', 'nice to meet you', 368, 'Mucho gusto.'),
    (7, 'igualmente', 'likewise', 369, 'Igualmente.'),
    (7, 'cuánto', 'how much', 370, '¿Cuánto cuesta?'),
    (7, 'cuántos', 'how many', 371, '¿Cuántos años tienes?'),
    (7, 'cómo', 'how', 372, '¿Cómo te llamas?'),
    (7, 'dónde', 'where', 373, '¿Dónde vives?'),
    (7, 'cuándo', 'when', 374, '¿Cuándo llegaste?'),
    (7, 'por qué', 'why', 375, '¿Por qué no vienes?'),
    (7, 'quién', 'who', 376, '¿Quién es?'),
    (7, 'cuál', 'which', 377, '¿Cuál prefieres?'),
    (7, 'nombre', 'name', 378, '¿Cuál es tu nombre?'),
    (7, 'edad', 'age', 379, '¿Qué edad tienes?'),
    (7, 'cumpleaños', 'birthday', 380, 'Hoy es mi cumpleaños.'),
    (7, 'dirección', 'address', 381, '¿Cuál es tu dirección?'),
    (7, 'correo', 'email, mail', 382, 'Te envío un correo.'),
    (7, 'cita', 'appointment, date', 383, 'Tengo una cita.'),
    (7, 'reunión', 'meeting', 384, 'Tengo una reunión.'),
    (7, 'vacaciones', 'vacation', 385, 'Estoy de vacaciones.'),
    (7, 'viaje', 'trip', 386, 'Buen viaje.'),
    (7, 'pasaporte', 'passport', 387, 'Necesito mi pasaporte.'),
    (7, 'boleto', 'ticket', 388, 'Compré el boleto.'),
    (7, 'equipaje', 'luggage', 389, 'Mi equipaje es pesado.'),
    (7, 'maleta', 'suitcase', 390, 'Hago la maleta.'),
    (7, 'llave', 'key', 391, '¿Dónde están las llaves?'),
    (7, 'cartera', 'wallet', 392, 'Olvidé mi cartera.'),
    (7, 'tarjeta', 'card', 393, 'Pago con tarjeta.'),
    (7, 'efectivo', 'cash', 394, 'Pago en efectivo.'),
    (7, 'cambio', 'change', 395, 'Aquí está tu cambio.'),
    (7, 'precio', 'price', 396, '¿Cuál es el precio?'),
    (7, 'barato', 'cheap', 397, 'Es muy barato.'),
    (7, 'caro', 'expensive', 398, 'Es muy caro.'),
    (7, 'gratis', 'free', 399, 'Es gratis.'),
    (7, 'abierto', 'open', 400, 'Está abierto.');

-- Island 8: Past Tense (Preterite forms)
INSERT OR IGNORE INTO cards (island_id, term, translation, frequency_rank, example_sentence) VALUES
    (8, 'fui', 'I went/was', 401, 'Fui al cine ayer.'),
    (8, 'fue', 'he/she/it went/was', 402, 'Fue una buena película.'),
    (8, 'tuve', 'I had', 403, 'Tuve un buen día.'),
    (8, 'tuvo', 'he/she had', 404, 'Tuvo suerte.'),
    (8, 'hice', 'I did/made', 405, 'Hice la tarea.'),
    (8, 'hizo', 'he/she did/made', 406, '¿Qué hizo?'),
    (8, 'dije', 'I said', 407, 'Te lo dije.'),
    (8, 'dijo', 'he/she said', 408, '¿Qué dijo?'),
    (8, 'pude', 'I could', 409, 'No pude ir.'),
    (8, 'puso', 'he/she put', 410, 'Puso la mesa.'),
    (8, 'vine', 'I came', 411, 'Vine temprano.'),
    (8, 'vino', 'he/she came', 412, 'Vino a visitarme.'),
    (8, 'quise', 'I wanted', 413, 'Quise ayudar.'),
    (8, 'quiso', 'he/she wanted', 414, 'No quiso venir.'),
    (8, 'supe', 'I found out', 415, 'Supe la verdad.'),
    (8, 'supo', 'he/she found out', 416, 'Supo todo.'),
    (8, 'estuve', 'I was', 417, 'Estuve enfermo.'),
    (8, 'estuvo', 'he/she was', 418, 'Estuvo aquí ayer.'),
    (8, 'comí', 'I ate', 419, 'Comí tacos.'),
    (8, 'comió', 'he/she ate', 420, 'Comió mucho.'),
    (8, 'bebí', 'I drank', 421, 'Bebí agua.'),
    (8, 'bebió', 'he/she drank', 422, 'Bebió café.'),
    (8, 'viví', 'I lived', 423, 'Viví en España.'),
    (8, 'vivió', 'he/she lived', 424, 'Vivió en México.'),
    (8, 'salí', 'I left', 425, 'Salí temprano.'),
    (8, 'salió', 'he/she left', 426, 'Salió a las ocho.'),
    (8, 'llegué', 'I arrived', 427, 'Llegué tarde.'),
    (8, 'llegó', 'he/she arrived', 428, 'Llegó a tiempo.'),
    (8, 'empecé', 'I started', 429, 'Empecé a estudiar.'),
    (8, 'empezó', 'he/she started', 430, 'Empezó a llover.'),
    (8, 'terminé', 'I finished', 431, 'Terminé el trabajo.'),
    (8, 'terminó', 'he/she finished', 432, 'Terminó la película.'),
    (8, 'pasé', 'I passed/spent', 433, 'Pasé un buen rato.'),
    (8, 'pasó', 'he/she passed/happened', 434, '¿Qué pasó?'),
    (8, 'encontré', 'I found', 435, 'Encontré las llaves.'),
    (8, 'encontró', 'he/she found', 436, 'Encontró trabajo.'),
    (8, 'pensé', 'I thought', 437, 'Pensé en ti.'),
    (8, 'pensó', 'he/she thought', 438, 'Pensó mucho.'),
    (8, 'compré', 'I bought', 439, 'Compré un regalo.'),
    (8, 'compró', 'he/she bought', 440, 'Compró una casa.'),
    (8, 'pagué', 'I paid', 441, 'Pagué la cuenta.'),
    (8, 'pagó', 'he/she paid', 442, 'Pagó con tarjeta.'),
    (8, 'trabajé', 'I worked', 443, 'Trabajé mucho.'),
    (8, 'trabajó', 'he/she worked', 444, 'Trabajó todo el día.'),
    (8, 'estudié', 'I studied', 445, 'Estudié español.'),
    (8, 'estudió', 'he/she studied', 446, 'Estudió medicina.'),
    (8, 'jugué', 'I played', 447, 'Jugué fútbol.'),
    (8, 'jugó', 'he/she played', 448, 'Jugó muy bien.'),
    (8, 'dormí', 'I slept', 449, 'Dormí ocho horas.'),
    (8, 'durmió', 'he/she slept', 450, 'Durmió toda la noche.');

-- Island 9: Subjunctive
INSERT OR IGNORE INTO cards (island_id, term, translation, frequency_rank, example_sentence) VALUES
    (9, 'quiero que', 'I want (that)', 451, 'Quiero que vengas.'),
    (9, 'espero que', 'I hope (that)', 452, 'Espero que estés bien.'),
    (9, 'ojalá', 'hopefully, I wish', 453, 'Ojalá venga.'),
    (9, 'es posible que', 'it''s possible that', 454, 'Es posible que llueva.'),
    (9, 'es necesario que', 'it''s necessary that', 455, 'Es necesario que estudies.'),
    (9, 'dudo que', 'I doubt that', 456, 'Dudo que sepa.'),
    (9, 'no creo que', 'I don''t think that', 457, 'No creo que sea verdad.'),
    (9, 'para que', 'so that', 458, 'Te lo digo para que sepas.'),
    (9, 'antes de que', 'before', 459, 'Antes de que te vayas.'),
    (9, 'cuando', 'when (future)', 460, 'Cuando llegues, llámame.'),
    (9, 'sea', 'be (subjunctive)', 461, 'Lo que sea.'),
    (9, 'tenga', 'have (subjunctive)', 462, 'Cuando tenga tiempo.'),
    (9, 'haga', 'do/make (subjunctive)', 463, 'Lo que haga falta.'),
    (9, 'pueda', 'can (subjunctive)', 464, 'Si pueda ayudar.'),
    (9, 'sepa', 'know (subjunctive)', 465, 'No creo que sepa.'),
    (9, 'vaya', 'go (subjunctive)', 466, 'Donde quiera que vaya.'),
    (9, 'venga', 'come (subjunctive)', 467, 'Espero que venga.'),
    (9, 'diga', 'say (subjunctive)', 468, 'Lo que diga.'),
    (9, 'quiera', 'want (subjunctive)', 469, 'Donde quiera.'),
    (9, 'esté', 'be (subjunctive)', 470, 'Espero que esté bien.'),
    (9, 'aunque sea', 'even if', 471, 'Aunque sea difícil.'),
    (9, 'como si', 'as if', 472, 'Habla como si supiera.'),
    (9, 'a menos que', 'unless', 473, 'A menos que llueva.'),
    (9, 'sin que', 'without', 474, 'Sin que lo sepa.'),
    (9, 'tal vez', 'maybe, perhaps', 475, 'Tal vez venga.'),
    (9, 'quizás', 'maybe, perhaps', 476, 'Quizás llegue tarde.'),
    (9, 'probablemente', 'probably', 477, 'Probablemente no venga.'),
    (9, 'necesito que', 'I need (that)', 478, 'Necesito que me ayudes.'),
    (9, 'prefiero que', 'I prefer (that)', 479, 'Prefiero que vengas.'),
    (9, 'sugiero que', 'I suggest (that)', 480, 'Sugiero que descanses.');

-- Island 6: Career & Work
INSERT OR IGNORE INTO cards (island_id, term, translation, frequency_rank, example_sentence) VALUES
    (6, 'jefe', 'boss', 481, 'Mi jefe es amable.'),
    (6, 'empleado', 'employee', 482, 'Soy empleado aquí.'),
    (6, 'empresa', 'company', 483, 'Trabajo en una empresa grande.'),
    (6, 'negocio', 'business', 484, 'Tengo un negocio.'),
    (6, 'cliente', 'client', 485, 'El cliente siempre tiene razón.'),
    (6, 'reunión', 'meeting', 486, 'Tenemos una reunión.'),
    (6, 'proyecto', 'project', 487, 'Trabajo en un proyecto.'),
    (6, 'informe', 'report', 488, 'Escribo el informe.'),
    (6, 'documento', 'document', 489, 'Firma el documento.'),
    (6, 'contrato', 'contract', 490, 'Firmé el contrato.'),
    (6, 'salario', 'salary', 491, 'Buen salario.'),
    (6, 'sueldo', 'wage', 492, '¿Cuánto es el sueldo?'),
    (6, 'vacaciones', 'vacation', 493, 'Necesito vacaciones.'),
    (6, 'experiencia', 'experience', 494, 'Tengo experiencia.'),
    (6, 'entrevista', 'interview', 495, 'Tengo una entrevista.'),
    (6, 'carrera', 'career', 496, 'Mi carrera profesional.'),
    (6, 'horario', 'schedule', 497, '¿Cuál es tu horario?'),
    (6, 'oficina', 'office', 498, 'Trabajo en una oficina.'),
    (6, 'computadora', 'computer', 499, 'Uso la computadora.'),
    (6, 'correo electrónico', 'email', 500, 'Te envío un correo electrónico.');

-- =============================================
-- SONG LESSONS TABLES
-- =============================================

-- Songs catalog
CREATE TABLE IF NOT EXISTS songs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    youtube_id TEXT UNIQUE,
    genius_id INTEGER,
    title TEXT NOT NULL,
    artist TEXT NOT NULL DEFAULT 'Bad Bunny',
    album TEXT,
    difficulty INTEGER DEFAULT 1 CHECK(difficulty IN (1, 2, 3)),
    duration_seconds INTEGER,
    thumbnail_url TEXT,
    audio_path TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Lyrics with timestamps
CREATE TABLE IF NOT EXISTS song_lines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    song_id INTEGER NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    line_number INTEGER NOT NULL,
    start_time_ms INTEGER NOT NULL,
    end_time_ms INTEGER NOT NULL,
    spanish_text TEXT NOT NULL,
    english_text TEXT NOT NULL,
    UNIQUE(song_id, line_number)
);

-- Key vocabulary from songs (links to flashcard system)
CREATE TABLE IF NOT EXISTS song_vocabulary (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    song_id INTEGER NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    card_id INTEGER REFERENCES cards(id) ON DELETE SET NULL,
    word TEXT NOT NULL,
    translation TEXT NOT NULL,
    is_key_vocab INTEGER DEFAULT 1,
    UNIQUE(song_id, word)
);

-- User progress per song (FSRS-based)
CREATE TABLE IF NOT EXISTS song_progress (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    song_id INTEGER NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    stability REAL DEFAULT 0,
    difficulty REAL DEFAULT 0,
    reps INTEGER DEFAULT 0,
    lapses INTEGER DEFAULT 0,
    state TEXT DEFAULT 'new' CHECK(state IN ('new', 'learning', 'review', 'relearning')),
    due DATETIME,
    last_review DATETIME,
    vocab_complete INTEGER DEFAULT 0,
    lyrics_complete INTEGER DEFAULT 0,
    listening_complete INTEGER DEFAULT 0,
    total_listens INTEGER DEFAULT 0,
    UNIQUE(user_id, song_id)
);

-- Session tracking for song lessons
CREATE TABLE IF NOT EXISTS song_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    song_id INTEGER NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    session_date DATE NOT NULL,
    mode TEXT NOT NULL CHECK(mode IN ('vocab', 'lyrics', 'listening', 'full')),
    vocab_reviewed INTEGER DEFAULT 0,
    vocab_correct INTEGER DEFAULT 0,
    lines_studied INTEGER DEFAULT 0,
    blanks_correct INTEGER DEFAULT 0,
    blanks_total INTEGER DEFAULT 0,
    xp_earned INTEGER DEFAULT 0,
    completed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for song tables
CREATE INDEX IF NOT EXISTS idx_song_lines_song ON song_lines(song_id);
CREATE INDEX IF NOT EXISTS idx_song_vocabulary_song ON song_vocabulary(song_id);
CREATE INDEX IF NOT EXISTS idx_song_progress_user ON song_progress(user_id);
CREATE INDEX IF NOT EXISTS idx_song_progress_due ON song_progress(due);
CREATE INDEX IF NOT EXISTS idx_song_sessions_user_date ON song_sessions(user_id, session_date);
