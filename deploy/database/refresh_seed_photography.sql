-- Refresh only our original sample dishes and placeholder thumbnails.
-- Unrelated recipes, uploaded images and edited seed titles are preserved.
BEGIN;
SELECT pg_advisory_xact_lock(hashtext('food-recipe-demo-seed-v1'));
DO $refresh$
DECLARE item JSONB; target_id INTEGER;
BEGIN
 FOR item IN SELECT value FROM jsonb_array_elements($dishes$[{"key": "v1/chechebsa", "title": "Skillet Shakshuka", "category": "Breakfast", "minutes": 30, "description": "Eggs gently baked in a warm tomato and pepper sauce. Bring the skillet to the table and scoop it up with good bread.", "image": "/images/recipes/shakshuka.webp", "ingredients": ["2 tablespoons olive oil", "1 onion, diced", "1 red bell pepper, diced", "2 garlic cloves, minced", "1 teaspoon ground cumin", "1 teaspoon sweet paprika", "400 g canned chopped tomatoes", "4 eggs", "Salt and black pepper to taste", "Fresh parsley and bread, to serve"], "steps": ["Warm the oil in a large skillet. Cook the onion and pepper until soft, then stir in garlic, cumin and paprika.", "Add the tomatoes and simmer gently for 10 to 15 minutes, until the sauce thickens. Season to taste.", "Make four small wells in the sauce and carefully crack an egg into each one.", "Cover and cook gently until the egg whites and yolks are set. Scatter over parsley and serve with bread."], "old_title": "Chechebsa with Spiced Butter"}, {"key": "v1/misir-wot", "title": "Tomato Lentil Soup", "category": "Dinner", "minutes": 35, "description": "A comforting bowl of red lentils, tomato and warm spices. A little lemon at the end makes the whole pot sing.", "image": "/images/recipes/lentil-soup.webp", "ingredients": ["1 tablespoon olive oil", "1 onion, chopped", "2 garlic cloves, minced", "1 carrot, diced", "1 teaspoon ground cumin", "1 cup red lentils, rinsed", "400 g canned chopped tomatoes", "4 cups vegetable stock", "1 tablespoon lemon juice", "Salt and black pepper to taste"], "steps": ["Warm the oil in a saucepan. Cook the onion and carrot for 5 minutes, then add garlic and cumin.", "Add the lentils, tomatoes and stock. Bring to a boil, then reduce to a gentle simmer.", "Cook for 20 to 25 minutes until the lentils and vegetables are tender. Blend a little of the soup for a creamier texture, if you like.", "Add lemon juice and season to taste. Serve warm with bread."], "old_title": "Misir Wot Red Lentil Stew"}, {"key": "v1/spiced-chickpeas", "title": "Lemon Hummus with Warm Pita", "category": "Snacks", "minutes": 15, "description": "Creamy chickpea hummus with lemon, tahini and a little olive oil. An easy bowl to put in the middle of the table.", "image": "/images/recipes/hummus.webp", "ingredients": ["400 g canned chickpeas, drained and rinsed", "3 tablespoons tahini", "2 tablespoons lemon juice", "1 small garlic clove", "2 tablespoons olive oil, plus a little to serve", "3 tablespoons cold water, plus more as needed", "1/2 teaspoon ground cumin", "Salt to taste", "Warm pita bread and chopped tomatoes, to serve"], "steps": ["Add chickpeas, tahini, lemon juice, garlic, cumin and salt to a food processor.", "Blend until smooth, adding the olive oil and cold water a little at a time until creamy.", "Taste and add more lemon or salt if needed. Spoon into a bowl and drizzle with olive oil.", "Serve with warm pita and chopped tomatoes."], "old_title": "Crispy Spiced Chickpeas"}]$dishes$::JSONB) LOOP
  SELECT r.id INTO target_id FROM recipe r JOIN public._food_recipe_seed s ON s.recipe_id=r.id
  WHERE s.seed_key=item->>'key' AND r.title=item->>'old_title' AND r.user_id='be146fbb-9379-4aac-8a48-870356fa5a8f';
  IF target_id IS NOT NULL THEN
   UPDATE recipe SET title=item->>'title',description=item->>'description',prep_time=(item->>'minutes')::INTEGER WHERE id=target_id;
   DELETE FROM recipe_ingredient WHERE recipe_id=target_id;
   INSERT INTO recipe_ingredient(recipe_id,ingredient) SELECT target_id,value FROM jsonb_array_elements_text(item->'ingredients');
   DELETE FROM recipe_step WHERE recipe_id=target_id;
   INSERT INTO recipe_step(recipe_id,step_no,step_description) SELECT target_id,ordinal::INTEGER,value FROM jsonb_array_elements_text(item->'steps') WITH ORDINALITY AS step(value,ordinal);
  END IF;
 END LOOP;
END;
$refresh$;
UPDATE "user" SET "FirstName"='Bowlup',"LastName"='Kitchen'
WHERE id='be146fbb-9379-4aac-8a48-870356fa5a8f' AND "FirstName"='Demo' AND "LastName"='Kitchen';
UPDATE recipe_image i SET img_url=photo.path
FROM public._food_recipe_seed s,recipe r,(VALUES
('v1/chechebsa','/images/recipes/shakshuka.webp'),
('v1/banana-pancakes','/images/recipes/banana-pancakes.webp'),
('v1/shiro-wot','/images/recipes/shiro.webp'),
('v1/tomato-pasta','/images/recipes/tomato-pasta.webp'),
('v1/misir-wot','/images/recipes/lentil-soup.webp'),
('v1/lemon-roast-chicken','/images/recipes/roast-chicken.webp'),
('v1/cardamom-rice-pudding','/images/recipes/rice-pudding.webp'),
('v1/baked-apples','/images/recipes/baked-apples.webp'),
('v1/spiced-chickpeas','/images/recipes/hummus.webp'),
('v1/avocado-tomato-toast','/images/recipes/avocado-toast.webp')) AS photo(seed_key,path)
WHERE s.seed_key=photo.seed_key AND s.recipe_id=i.recipe_id AND r.id=i.recipe_id
 AND r.user_id='be146fbb-9379-4aac-8a48-870356fa5a8f'
 AND (i.img_url LIKE '/img/%' OR i.img_url LIKE '/images/recipes/%');
COMMIT;
