-- Migration: Add audio_path, genius_id, and album columns to songs table
-- Note: These columns now exist in schema.sql, so we skip the ALTER TABLE statements
-- The columns are: genius_id INTEGER, album TEXT, audio_path TEXT

-- Seed songs from the songs folder (Bad Bunny - Un Verano Sin Ti)
INSERT OR IGNORE INTO songs (title, artist, album, difficulty, audio_path) VALUES
    ('Moscow Mule', 'Bad Bunny', 'Un Verano Sin Ti', 1, '01. Moscow Mule.mp3'),
    ('Después de la Playa', 'Bad Bunny', 'Un Verano Sin Ti', 2, '02. Después de la Playa.mp3'),
    ('Me Porto Bonito', 'Bad Bunny', 'Un Verano Sin Ti', 1, '03. Me Porto Bonito.mp3'),
    ('Tití Me Preguntó', 'Bad Bunny', 'Un Verano Sin Ti', 2, '04. Tití Me Preguntó.mp3'),
    ('Un Ratito', 'Bad Bunny', 'Un Verano Sin Ti', 2, '05. Un Ratito.mp3'),
    ('Yo No Soy Celoso', 'Bad Bunny', 'Un Verano Sin Ti', 2, '06. Yo No Soy Celoso.mp3'),
    ('Tarot', 'Bad Bunny', 'Un Verano Sin Ti', 2, '07. Tarot.mp3'),
    ('Neverita', 'Bad Bunny', 'Un Verano Sin Ti', 1, '08. Neverita.mp3'),
    ('La Corriente', 'Bad Bunny', 'Un Verano Sin Ti', 2, '09. La Corriente.mp3'),
    ('Efecto', 'Bad Bunny', 'Un Verano Sin Ti', 2, '10. Efecto.mp3'),
    ('Party', 'Bad Bunny', 'Un Verano Sin Ti', 1, '11. Party.mp3'),
    ('Aguacero', 'Bad Bunny', 'Un Verano Sin Ti', 2, '12. Aguacero.mp3'),
    ('Enséñame a Bailar', 'Bad Bunny', 'Un Verano Sin Ti', 2, '13. Enséñame a Bailar.mp3'),
    ('Ojitos Lindos', 'Bad Bunny', 'Un Verano Sin Ti', 1, '14. Ojitos Lindos.mp3'),
    ('Dos Mil 16', 'Bad Bunny', 'Un Verano Sin Ti', 2, '15. Dos Mil 16.mp3'),
    ('El Apagón', 'Bad Bunny', 'Un Verano Sin Ti', 3, '16. El Apagón.mp3'),
    ('Otro Atardecer', 'Bad Bunny', 'Un Verano Sin Ti', 2, '17. Otro Atardecer.mp3'),
    ('Un Coco', 'Bad Bunny', 'Un Verano Sin Ti', 2, '18. Un Coco.mp3'),
    ('Andrea', 'Bad Bunny', 'Un Verano Sin Ti', 2, '19. Andrea.mp3'),
    ('Me Fui de Vacaciones', 'Bad Bunny', 'Un Verano Sin Ti', 2, '20. Me Fui de Vacaciones.mp3'),
    ('Un Verano Sin Ti', 'Bad Bunny', 'Un Verano Sin Ti', 2, '21. Un Verano Sin Ti.mp3'),
    ('Agosto', 'Bad Bunny', 'Un Verano Sin Ti', 2, '22. Agosto.mp3'),
    ('Callaita', 'Bad Bunny', 'Un Verano Sin Ti', 1, '23. Callaita.mp3');

-- Bad Bunny - DeBÍ TiRAR MáS FOToS
INSERT OR IGNORE INTO songs (title, artist, album, difficulty, audio_path) VALUES
    ('NUEVAYoL', 'Bad Bunny', 'DeBÍ TiRAR MáS FOToS', 2, '01. NUEVAYoL.flac'),
    ('VOY A LLeVARTE PA PR', 'Bad Bunny', 'DeBÍ TiRAR MáS FOToS', 2, '02. VOY A LLeVARTE PA PR.flac'),
    ('BAILE INoLVIDABLE', 'Bad Bunny', 'DeBÍ TiRAR MáS FOToS', 2, '03. BAILE INoLVIDABLE.flac'),
    ('PERFuMITO NUEVO', 'Bad Bunny', 'DeBÍ TiRAR MáS FOToS', 2, '04. PERFuMITO NUEVO.flac'),
    ('WELTiTA', 'Bad Bunny', 'DeBÍ TiRAR MáS FOToS', 2, '05. WELTiTA.flac'),
    ('VeLDÁ', 'Bad Bunny', 'DeBÍ TiRAR MáS FOToS', 2, '06. VeLDÁ.flac'),
    ('EL CLúB', 'Bad Bunny', 'DeBÍ TiRAR MáS FOToS', 2, '07. EL CLúB.flac'),
    ('KETU TeCRÉ', 'Bad Bunny', 'DeBÍ TiRAR MáS FOToS', 2, '08. KETU TeCRÉ.flac'),
    ('BOKeTE', 'Bad Bunny', 'DeBÍ TiRAR MáS FOToS', 2, '09. BOKeTE.flac'),
    ('KLOuFRENS', 'Bad Bunny', 'DeBÍ TiRAR MáS FOToS', 2, '10. KLOuFRENS.flac'),
    ('TURiSTA', 'Bad Bunny', 'DeBÍ TiRAR MáS FOToS', 2, '11. TURiSTA.flac'),
    ('CAFé CON RON', 'Bad Bunny', 'DeBÍ TiRAR MáS FOToS', 2, '12. CAFé CON RON.flac'),
    ('PIToRRO DE COCO', 'Bad Bunny', 'DeBÍ TiRAR MáS FOToS', 2, '13. PIToRRO DE COCO.flac'),
    ('LO QUE LE PASÓ A HAWAii', 'Bad Bunny', 'DeBÍ TiRAR MáS FOToS', 2, '14. LO QUE LE PASÓ A HAWAii.flac'),
    ('EoO', 'Bad Bunny', 'DeBÍ TiRAR MáS FOToS', 2, '15. EoO.flac'),
    ('DtMF', 'Bad Bunny', 'DeBÍ TiRAR MáS FOToS', 2, '16. DtMF.flac'),
    ('LA MuDANZA', 'Bad Bunny', 'DeBÍ TiRAR MáS FOToS', 2, '17. LA MuDANZA.flac');

-- Bad Bunny - nadie sabe lo que va a pasar mañana
INSERT OR IGNORE INTO songs (title, artist, album, difficulty, audio_path) VALUES
    ('NADIE SABE', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '01. NADIE SABE.mp3'),
    ('MONACO', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '02. MONACO.mp3'),
    ('FINA', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '03. FINA.mp3'),
    ('HIBIKI', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '04. HIBIKI.mp3'),
    ('MR. OCTOBER', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '05. MR. OCTOBER.mp3'),
    ('CYBERTRUCK', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '06. CYBERTRUCK.mp3'),
    ('VOU 787', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '07. VOU 787.mp3'),
    ('SEDA', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '08. SEDA.mp3'),
    ('GRACIAS POR NADA', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '09. GRACIAS POR NADA.mp3'),
    ('TELEFONO NUEVO', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '10. TELEFONO NUEVO.mp3'),
    ('BABY NUEVA', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '11. BABY NUEVA.mp3'),
    ('MERCEDES CAROTA', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '12. MERCEDES CAROTA.mp3'),
    ('LOS PITS', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '13. LOS PITS.mp3'),
    ('VUELVE CANDY B', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '14. VUELVE CANDY B.mp3'),
    ('BATICANO', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '15. BATICANO.mp3'),
    ('NO ME QUIERO CASAR', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '16. NO ME QUIERO CASAR.mp3'),
    ('WHERE SHE GOES', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 1, '17. WHERE SHE GOES.mp3'),
    ('THUNDER Y LIGHTNING', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '18. THUNDER Y LIGHTNING.mp3'),
    ('PERRO NEGRO', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '19. PERRO NEGRO.mp3'),
    ('EUROPA _', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '20. EUROPA _(.mp3'),
    ('ACHO PR', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '21. ACHO PR.mp3'),
    ('UN PREVIEW', 'Bad Bunny', 'nadie sabe lo que va a pasar mañana', 2, '22. UN PREVIEW.mp3');
