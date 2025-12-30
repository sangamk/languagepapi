-- Seed data for Bad Bunny song lessons
-- Run this after schema.sql to populate initial songs

-- ============================================
-- Song 1: Callaíta (Beginner - slow tempo, clear pronunciation)
-- ============================================
INSERT OR IGNORE INTO songs (youtube_id, title, artist, difficulty, duration_seconds, thumbnail_url)
VALUES ('FxQTY-W6GIo', 'Callaíta', 'Bad Bunny', 1, 251, 'https://i.ytimg.com/vi/FxQTY-W6GIo/maxresdefault.jpg');

-- Lyrics with timestamps (selected verses for learning)
INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'),
    1, 12000, 16000, 'Ella es callaíta, pero pa'' montar es una fiera', 'She''s quiet, but for riding she''s a beast'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND line_number = 1);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'),
    2, 16000, 20000, 'Ninguna se le compara', 'No one compares to her'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND line_number = 2);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'),
    3, 20000, 24000, 'Tiene la nota y también tiene su carrera', 'She has the vibe and also has her career'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND line_number = 3);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'),
    4, 24000, 28000, 'Un sol en la noche, la luna la espera', 'A sun in the night, the moon waits for her'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND line_number = 4);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'),
    5, 32000, 36000, 'Callaíta, pero los domingos en la disco grita', 'Quiet, but on Sundays in the club she screams'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND line_number = 5);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'),
    6, 36000, 40000, 'Tra, tra, tra, tra, callaíta', 'Tra, tra, tra, tra, quiet one'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND line_number = 6);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'),
    7, 40000, 44000, 'Siempre callaíta, tú eres mi favorita', 'Always quiet, you are my favorite'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND line_number = 7);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'),
    8, 48000, 52000, 'Bebe, no puedo dejar de mirarte', 'Baby, I can''t stop looking at you'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND line_number = 8);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'),
    9, 52000, 56000, 'Si estás solita, te puedo acompañar', 'If you''re alone, I can keep you company'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND line_number = 9);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'),
    10, 56000, 60000, 'El problema es que ya me tiene enamorado', 'The problem is she already has me in love'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND line_number = 10);

-- Key vocabulary from Callaíta
INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'), 'callaíta', 'quiet girl (diminutive)', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND word = 'callaíta');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'), 'montar', 'to ride / to mount', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND word = 'montar');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'), 'fiera', 'beast / fierce one', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND word = 'fiera');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'), 'carrera', 'career / race', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND word = 'carrera');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'), 'luna', 'moon', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND word = 'luna');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'), 'noche', 'night', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND word = 'noche');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'), 'disco', 'club / disco', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND word = 'disco');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'), 'favorita', 'favorite (feminine)', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND word = 'favorita');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'), 'solita', 'alone (feminine diminutive)', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND word = 'solita');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo'), 'enamorado', 'in love', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'FxQTY-W6GIo') AND word = 'enamorado');

-- ============================================
-- Song 2: Dakiti (Beginner - catchy, repetitive chorus)
-- ============================================
INSERT OR IGNORE INTO songs (youtube_id, title, artist, difficulty, duration_seconds, thumbnail_url)
VALUES ('TmKh7lAwnBI', 'Dakiti', 'Bad Bunny ft. Jhay Cortez', 1, 205, 'https://i.ytimg.com/vi/TmKh7lAwnBI/maxresdefault.jpg');

-- Lyrics with timestamps
INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'),
    1, 15000, 19000, 'Dime si te quedas o si te vas', 'Tell me if you''re staying or if you''re leaving'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND line_number = 1);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'),
    2, 19000, 23000, 'Si lo hacemos lento o lo hacemos ya', 'If we do it slow or we do it now'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND line_number = 2);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'),
    3, 23000, 27000, 'Me tiene portándome mal', 'She has me behaving badly'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND line_number = 3);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'),
    4, 27000, 31000, 'Todos los días a mí me dan ganas de verte', 'Every day I feel like seeing you'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND line_number = 4);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'),
    5, 35000, 39000, 'Pero si me porto bien, quizás me lleve pa'' PR', 'But if I behave well, maybe she''ll take me to PR'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND line_number = 5);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'),
    6, 39000, 43000, 'Ella es la dura de las duras', 'She''s the baddest of the bad'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND line_number = 6);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'),
    7, 43000, 47000, 'Tiene la cara de bebé, pero es madura', 'She has a baby face, but she''s mature'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND line_number = 7);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'),
    8, 51000, 55000, 'Yo sé que tú quieres conmigo', 'I know that you want to be with me'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND line_number = 8);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'),
    9, 55000, 59000, 'No tienes que decirlo', 'You don''t have to say it'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND line_number = 9);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'),
    10, 59000, 63000, 'Si ya te conozco', 'I already know you'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND line_number = 10);

-- Key vocabulary from Dakiti
INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'), 'quedarse', 'to stay', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND word = 'quedarse');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'), 'lento', 'slow', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND word = 'lento');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'), 'portarse', 'to behave', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND word = 'portarse');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'), 'ganas', 'desire / urge', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND word = 'ganas');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'), 'dura', 'tough / baddie (slang)', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND word = 'dura');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'), 'cara', 'face', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND word = 'cara');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'), 'madura', 'mature', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND word = 'madura');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'), 'conmigo', 'with me', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND word = 'conmigo');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'), 'decir', 'to say / to tell', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND word = 'decir');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI'), 'conocer', 'to know (someone)', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'TmKh7lAwnBI') AND word = 'conocer');

-- ============================================
-- Song 3: Tití Me Preguntó (Intermediate - faster, more vocabulary)
-- ============================================
INSERT OR IGNORE INTO songs (youtube_id, title, artist, difficulty, duration_seconds, thumbnail_url)
VALUES ('OmHgU6hAY3s', 'Tití Me Preguntó', 'Bad Bunny', 2, 241, 'https://i.ytimg.com/vi/OmHgU6hAY3s/maxresdefault.jpg');

-- Lyrics with timestamps
INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'),
    1, 8000, 12000, 'Mi tití me preguntó si tengo muchas novias', 'My auntie asked me if I have many girlfriends'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND line_number = 1);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'),
    2, 12000, 16000, 'Muchas novias, muchas novias', 'Many girlfriends, many girlfriends'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND line_number = 2);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'),
    3, 16000, 20000, 'Y yo le dije que sí, que ando con varias', 'And I told her yes, that I''m with several'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND line_number = 3);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'),
    4, 20000, 24000, 'Varias baby, una para cada día', 'Several baby, one for each day'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND line_number = 4);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'),
    5, 28000, 32000, 'La de lunes ya tiene a alguien', 'The Monday one already has someone'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND line_number = 5);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'),
    6, 32000, 36000, 'La de martes está ready pa'' irse de viaje', 'The Tuesday one is ready to go on a trip'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND line_number = 6);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'),
    7, 36000, 40000, 'La de miércoles quiere aprender de mi lenguaje', 'The Wednesday one wants to learn my language'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND line_number = 7);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'),
    8, 40000, 44000, 'Y la de jueves me tiene loco, no tiene iguales', 'And the Thursday one drives me crazy, she has no equal'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND line_number = 8);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'),
    9, 48000, 52000, 'La de viernes tiene todas las ganas', 'The Friday one has all the desire'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND line_number = 9);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'),
    10, 52000, 56000, 'La del sábado no sale de mi cama', 'The Saturday one doesn''t leave my bed'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND line_number = 10);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'),
    11, 56000, 60000, 'Y los domingos tengo libre, así que llama', 'And Sundays I have free, so call me'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND line_number = 11);

INSERT OR IGNORE INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
SELECT
    (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'),
    12, 60000, 64000, 'Ya no quepo en Instagram de tanta fama', 'I don''t fit on Instagram anymore from so much fame'
WHERE NOT EXISTS (SELECT 1 FROM song_lines WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND line_number = 12);

-- Key vocabulary from Tití Me Preguntó
INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'), 'tití', 'auntie (Puerto Rican slang)', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND word = 'tití');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'), 'preguntar', 'to ask', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND word = 'preguntar');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'), 'novias', 'girlfriends', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND word = 'novias');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'), 'varias', 'several / various', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND word = 'varias');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'), 'lunes', 'Monday', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND word = 'lunes');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'), 'martes', 'Tuesday', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND word = 'martes');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'), 'miércoles', 'Wednesday', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND word = 'miércoles');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'), 'jueves', 'Thursday', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND word = 'jueves');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'), 'viernes', 'Friday', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND word = 'viernes');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'), 'sábado', 'Saturday', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND word = 'sábado');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'), 'domingo', 'Sunday', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND word = 'domingo');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'), 'cama', 'bed', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND word = 'cama');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'), 'fama', 'fame', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND word = 'fama');

INSERT OR IGNORE INTO song_vocabulary (song_id, word, translation, is_key_vocab)
SELECT (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s'), 'lenguaje', 'language', 1
WHERE NOT EXISTS (SELECT 1 FROM song_vocabulary WHERE song_id = (SELECT id FROM songs WHERE youtube_id = 'OmHgU6hAY3s') AND word = 'lenguaje');
