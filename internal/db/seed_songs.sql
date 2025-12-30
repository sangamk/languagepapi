-- Seed data for Bad Bunny song lessons
-- Run this after schema.sql to populate initial songs

-- ============================================
-- Song 1: Callaíta (Beginner - slow tempo, clear pronunciation)
-- ============================================
INSERT INTO songs (youtube_id, title, artist, difficulty, duration_seconds, thumbnail_url)
VALUES ('FxQTY-W6GIo', 'Callaíta', 'Bad Bunny', 1, 251, 'https://i.ytimg.com/vi/FxQTY-W6GIo/maxresdefault.jpg');

-- Lyrics with timestamps (selected verses for learning)
INSERT INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text) VALUES
(1, 1, 12000, 16000, 'Ella es callaíta, pero pa'' montar es una fiera', 'She''s quiet, but for riding she''s a beast'),
(1, 2, 16000, 20000, 'Ninguna se le compara', 'No one compares to her'),
(1, 3, 20000, 24000, 'Tiene la nota y también tiene su carrera', 'She has the vibe and also has her career'),
(1, 4, 24000, 28000, 'Un sol en la noche, la luna la espera', 'A sun in the night, the moon waits for her'),
(1, 5, 32000, 36000, 'Callaíta, pero los domingos en la disco grita', 'Quiet, but on Sundays in the club she screams'),
(1, 6, 36000, 40000, 'Tra, tra, tra, tra, callaíta', 'Tra, tra, tra, tra, quiet one'),
(1, 7, 40000, 44000, 'Siempre callaíta, tú eres mi favorita', 'Always quiet, you are my favorite'),
(1, 8, 48000, 52000, 'Bebe, no puedo dejar de mirarte', 'Baby, I can''t stop looking at you'),
(1, 9, 52000, 56000, 'Si estás solita, te puedo acompañar', 'If you''re alone, I can keep you company'),
(1, 10, 56000, 60000, 'El problema es que ya me tiene enamorado', 'The problem is she already has me in love');

-- Key vocabulary from Callaíta
INSERT INTO song_vocabulary (song_id, word, translation, is_key_vocab) VALUES
(1, 'callaíta', 'quiet girl (diminutive)', 1),
(1, 'montar', 'to ride / to mount', 1),
(1, 'fiera', 'beast / fierce one', 1),
(1, 'carrera', 'career / race', 1),
(1, 'luna', 'moon', 1),
(1, 'noche', 'night', 1),
(1, 'disco', 'club / disco', 1),
(1, 'favorita', 'favorite (feminine)', 1),
(1, 'solita', 'alone (feminine diminutive)', 1),
(1, 'enamorado', 'in love', 1);

-- ============================================
-- Song 2: Dakiti (Beginner - catchy, repetitive chorus)
-- ============================================
INSERT INTO songs (youtube_id, title, artist, difficulty, duration_seconds, thumbnail_url)
VALUES ('TmKh7lAwnBI', 'Dakiti', 'Bad Bunny ft. Jhay Cortez', 1, 205, 'https://i.ytimg.com/vi/TmKh7lAwnBI/maxresdefault.jpg');

-- Lyrics with timestamps
INSERT INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text) VALUES
(2, 1, 15000, 19000, 'Dime si te quedas o si te vas', 'Tell me if you''re staying or if you''re leaving'),
(2, 2, 19000, 23000, 'Si lo hacemos lento o lo hacemos ya', 'If we do it slow or we do it now'),
(2, 3, 23000, 27000, 'Me tiene portándome mal', 'She has me behaving badly'),
(2, 4, 27000, 31000, 'Todos los días a mí me dan ganas de verte', 'Every day I feel like seeing you'),
(2, 5, 35000, 39000, 'Pero si me porto bien, quizás me lleve pa'' PR', 'But if I behave well, maybe she''ll take me to PR'),
(2, 6, 39000, 43000, 'Ella es la dura de las duras', 'She''s the baddest of the bad'),
(2, 7, 43000, 47000, 'Tiene la cara de bebé, pero es madura', 'She has a baby face, but she''s mature'),
(2, 8, 51000, 55000, 'Yo sé que tú quieres conmigo', 'I know that you want to be with me'),
(2, 9, 55000, 59000, 'No tienes que decirlo', 'You don''t have to say it'),
(2, 10, 59000, 63000, 'Si ya te conozco', 'I already know you');

-- Key vocabulary from Dakiti
INSERT INTO song_vocabulary (song_id, word, translation, is_key_vocab) VALUES
(2, 'quedarse', 'to stay', 1),
(2, 'lento', 'slow', 1),
(2, 'portarse', 'to behave', 1),
(2, 'ganas', 'desire / urge', 1),
(2, 'dura', 'tough / baddie (slang)', 1),
(2, 'cara', 'face', 1),
(2, 'madura', 'mature', 1),
(2, 'conmigo', 'with me', 1),
(2, 'decir', 'to say / to tell', 1),
(2, 'conocer', 'to know (someone)', 1);

-- ============================================
-- Song 3: Tití Me Preguntó (Intermediate - faster, more vocabulary)
-- ============================================
INSERT INTO songs (youtube_id, title, artist, difficulty, duration_seconds, thumbnail_url)
VALUES ('OmHgU6hAY3s', 'Tití Me Preguntó', 'Bad Bunny', 2, 241, 'https://i.ytimg.com/vi/OmHgU6hAY3s/maxresdefault.jpg');

-- Lyrics with timestamps
INSERT INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text) VALUES
(3, 1, 8000, 12000, 'Mi tití me preguntó si tengo muchas novias', 'My auntie asked me if I have many girlfriends'),
(3, 2, 12000, 16000, 'Muchas novias, muchas novias', 'Many girlfriends, many girlfriends'),
(3, 3, 16000, 20000, 'Y yo le dije que sí, que ando con varias', 'And I told her yes, that I''m with several'),
(3, 4, 20000, 24000, 'Varias baby, una para cada día', 'Several baby, one for each day'),
(3, 5, 28000, 32000, 'La de lunes ya tiene a alguien', 'The Monday one already has someone'),
(3, 6, 32000, 36000, 'La de martes está ready pa'' irse de viaje', 'The Tuesday one is ready to go on a trip'),
(3, 7, 36000, 40000, 'La de miércoles quiere aprender de mi lenguaje', 'The Wednesday one wants to learn my language'),
(3, 8, 40000, 44000, 'Y la de jueves me tiene loco, no tiene iguales', 'And the Thursday one drives me crazy, she has no equal'),
(3, 9, 48000, 52000, 'La de viernes tiene todas las ganas', 'The Friday one has all the desire'),
(3, 10, 52000, 56000, 'La del sábado no sale de mi cama', 'The Saturday one doesn''t leave my bed'),
(3, 11, 56000, 60000, 'Y los domingos tengo libre, así que llama', 'And Sundays I have free, so call me'),
(3, 12, 60000, 64000, 'Ya no quepo en Instagram de tanta fama', 'I don''t fit on Instagram anymore from so much fame');

-- Key vocabulary from Tití Me Preguntó
INSERT INTO song_vocabulary (song_id, word, translation, is_key_vocab) VALUES
(3, 'tití', 'auntie (Puerto Rican slang)', 1),
(3, 'preguntar', 'to ask', 1),
(3, 'novias', 'girlfriends', 1),
(3, 'varias', 'several / various', 1),
(3, 'lunes', 'Monday', 1),
(3, 'martes', 'Tuesday', 1),
(3, 'miércoles', 'Wednesday', 1),
(3, 'jueves', 'Thursday', 1),
(3, 'viernes', 'Friday', 1),
(3, 'sábado', 'Saturday', 1),
(3, 'domingo', 'Sunday', 1),
(3, 'cama', 'bed', 1),
(3, 'fama', 'fame', 1),
(3, 'lenguaje', 'language', 1);
