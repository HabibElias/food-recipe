-- Optional sample content. Existing accounts, recipes and edits are preserved.
-- Run after the app schema/bootstrap. Repeating this transaction adds no duplicates.
BEGIN;
SELECT pg_advisory_xact_lock(hashtext('food-recipe-demo-seed-v1'));

CREATE TABLE IF NOT EXISTS public._food_recipe_seed (
    seed_key TEXT PRIMARY KEY,
    recipe_id INTEGER NOT NULL UNIQUE REFERENCES recipe(id) ON DELETE CASCADE,
    seeded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $seed$
DECLARE
    author_id CONSTANT UUID := 'be146fbb-9379-4aac-8a48-870356fa5a8f';
    item JSONB;
    new_recipe_id INTEGER;
BEGIN
    IF EXISTS (SELECT 1 FROM "user" WHERE id = author_id AND
        (email <> 'demo-kitchen@example.invalid' OR username <> 'demo_kitchen')) THEN
        RAISE EXCEPTION 'The demo identity is already used; no seed changes applied';
    END IF;

    -- Generate and hash an undisclosed random password. This is a sample author,
    -- not a shared demo login or an administrator account.
    INSERT INTO "user" (id, email, password, username, "FirstName", "LastName", bio)
    VALUES (author_id, 'demo-kitchen@example.invalid',
        crypt(encode(gen_random_bytes(32), 'hex'), gen_salt('bf', 10)),
        'demo_kitchen', 'Demo', 'Kitchen',
        'Sample recipes to help you explore the app. Create your own account to share your cooking.')
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO category (category_name, category_description) VALUES
        ('Breakfast', 'Simple recipes for a good start to the day'),
        ('Lunch', 'Everyday meals for your midday break'),
        ('Dinner', 'Comforting meals to gather around'),
        ('Dessert', 'Sweet treats to finish a meal'),
        ('Snacks', 'Quick bites and small plates')
    ON CONFLICT (category_name) DO NOTHING;

    FOR item IN SELECT value FROM jsonb_array_elements($recipes$[
      {
        "key":"v1/chechebsa", "title":"Chechebsa with Spiced Butter", "category":"Breakfast", "minutes":30,
        "description":"Torn skillet flatbread tossed with warm berbere and butter, inspired by an Ethiopian breakfast favorite.",
        "image":"/img/breakfast.svg",
        "ingredients":["1 cup all-purpose flour","3/4 cup water, plus more if needed","1/4 teaspoon salt","2 tablespoons butter or niter kibbeh","1 teaspoon berbere, adjusted to taste","Plain yogurt, to serve"],
        "steps":["Whisk flour, salt and water into a smooth, thick batter. Rest for 5 minutes.","Heat a lightly greased skillet over medium heat. Spread the batter into a thin flatbread and cook until firm and lightly browned on both sides.","Tear the cooked bread into bite-sized pieces.","Melt the butter over low heat, stir in berbere briefly, then toss in the bread until evenly coated. Serve warm with yogurt."]
      },
      {
        "key":"v1/banana-pancakes", "title":"Fluffy Banana Pancakes", "category":"Breakfast", "minutes":20,
        "description":"Soft banana pancakes made with pantry staples. A relaxed weekend breakfast with fresh fruit.",
        "image":"/img/breakfast.svg",
        "ingredients":["1 ripe banana","1 egg","3/4 cup milk","1 cup all-purpose flour","2 teaspoons baking powder","1 tablespoon sugar","Pinch of salt","Butter or oil for the pan"],
        "steps":["Mash the banana in a bowl and whisk in the egg and milk.","Mix flour, baking powder, sugar and salt separately, then fold into the banana mixture just until combined.","Heat a lightly greased skillet over medium heat. Add small ladles of batter.","Cook until bubbles form and the edges set, then flip and cook until the centers are fully set. Serve with fruit."]
      },
      {
        "key":"v1/shiro-wot", "title":"Weeknight Shiro Wot", "category":"Lunch", "minutes":30,
        "description":"A smooth chickpea-flour stew with onion and warm spices. Serve with injera or rice for a filling lunch.",
        "image":"/img/lunch.svg",
        "ingredients":["1 medium onion, finely chopped","2 tablespoons oil","2 garlic cloves, minced","1 teaspoon berbere, or to taste","1 small tomato, finely chopped","1/2 cup shiro powder or chickpea flour","2 cups water, plus more as needed","Salt to taste","Injera or cooked rice, to serve"],
        "steps":["Cook the onion in oil over medium-low heat until soft. Add garlic and berbere and stir for 30 seconds.","Add the tomato and cook until it breaks down.","Whisk the shiro powder with a little cool water to make a smooth paste. Gradually whisk it into the pan with the remaining water.","Simmer gently for 15 to 20 minutes, stirring often, until smooth and cooked through. Add water to adjust the thickness and season with salt. Serve warm."]
      },
      {
        "key":"v1/tomato-pasta", "title":"One-Pot Tomato and Basil Pasta", "category":"Lunch", "minutes":25,
        "description":"Pasta simmered in a tomato sauce with garlic and basil, with just one pot to wash afterward.",
        "image":"/img/lunch.svg",
        "ingredients":["200 g dried pasta","1 tablespoon olive oil","2 garlic cloves, minced","400 g canned chopped tomatoes","2 cups vegetable stock, plus more if needed","1/2 teaspoon dried oregano","A handful of fresh basil","Salt and black pepper to taste"],
        "steps":["Warm olive oil in a large pot and gently cook garlic for 30 seconds.","Add tomatoes, stock and oregano. Bring to a simmer.","Add pasta and cook uncovered, stirring frequently, until tender. Add more hot water or stock if the pot becomes dry before the pasta cooks.","Season to taste and fold in torn basil just before serving."]
      },
      {
        "key":"v1/misir-wot", "title":"Misir Wot Red Lentil Stew", "category":"Dinner", "minutes":45,
        "description":"A comforting Ethiopian-inspired red lentil stew with a rich onion base and a gently warming berbere sauce.",
        "image":"/img/dinner.svg",
        "ingredients":["1 cup red lentils, rinsed","1 large onion, finely chopped","2 tablespoons oil","2 garlic cloves, minced","1 teaspoon grated ginger","1 tablespoon berbere, adjusted to taste","1 tablespoon tomato paste","3 cups water or vegetable stock","Salt to taste"],
        "steps":["Cook the onion in oil over medium-low heat until soft and lightly golden.","Add garlic, ginger, berbere and tomato paste. Stir for 1 minute.","Add the lentils and water or stock. Bring to a gentle simmer.","Cook for 25 to 30 minutes, stirring and adding water as needed, until the lentils are tender and creamy. Season and serve with injera or rice."]
      },
      {
        "key":"v1/lemon-roast-chicken", "title":"Lemon and Herb Roast Chicken", "category":"Dinner", "minutes":90,
        "description":"A lemony roast chicken with garlic, herbs and onions. A simple centerpiece for a shared dinner.",
        "image":"/img/CImage.jpg",
        "ingredients":["1 whole chicken, about 1.5 kg","2 tablespoons olive oil","1 lemon, halved","4 garlic cloves, crushed","1 teaspoon dried thyme","1 onion, cut into wedges","Salt and black pepper to taste"],
        "steps":["Heat the oven to 200 C. Place the onion wedges in a roasting dish. Pat the chicken dry with paper towels; do not rinse raw poultry.","Rub the chicken with oil, thyme, salt and pepper. Place lemon and garlic inside the cavity, then set the chicken on the onions.","Roast for about 70 to 85 minutes, adjusting for size, until a thermometer reads at least 74 C in the thickest breast and inner thigh without touching bone.","Rest for 10 minutes before carving. Serve with the roasted onions and pan juices."]
      },
      {
        "key":"v1/cardamom-rice-pudding", "title":"Cardamom Rice Pudding", "category":"Dessert", "minutes":35,
        "description":"Creamy rice pudding with cardamom and vanilla. Enjoy it warm or chilled after dinner.",
        "image":"/img/dessert.svg",
        "ingredients":["1/2 cup short-grain rice, rinsed","2 cups milk","1 cup water","3 tablespoons sugar","1/4 teaspoon ground cardamom","1/2 teaspoon vanilla extract","A pinch of salt"],
        "steps":["Combine the rice, water and salt in a saucepan. Simmer gently until most of the water is absorbed.","Stir in milk, sugar and cardamom.","Cook over low heat, stirring frequently, until the rice is tender and the mixture is creamy, about 20 minutes. Add a little milk if needed.","Stir in vanilla and serve warm, or cool promptly and refrigerate for a chilled dessert."]
      },
      {
        "key":"v1/baked-apples", "title":"Cinnamon Baked Apples", "category":"Dessert", "minutes":35,
        "description":"Tender baked apples with oats, cinnamon and a little honey. Serve with yogurt or ice cream.",
        "image":"/img/dessert.svg",
        "ingredients":["2 large apples","3 tablespoons rolled oats","1 tablespoon softened butter","1 tablespoon honey or maple syrup","1/2 teaspoon ground cinnamon","2 tablespoons water"],
        "steps":["Heat the oven to 190 C. Halve the apples and carefully remove the cores.","Mix oats, butter, honey and cinnamon. Spoon the mixture into the apple centers.","Place apples in a small baking dish and pour the water around them.","Bake for 25 to 30 minutes until tender. Let cool slightly before serving."]
      },
      {
        "key":"v1/spiced-chickpeas", "title":"Crispy Spiced Chickpeas", "category":"Snacks", "minutes":35,
        "description":"Roasted chickpeas with paprika and cumin, ready for snacking or adding crunch to a salad.",
        "image":"/img/quickbite.svg",
        "ingredients":["400 g canned chickpeas, drained and rinsed","1 tablespoon olive oil","1/2 teaspoon ground cumin","1/2 teaspoon smoked paprika","1/4 teaspoon salt"],
        "steps":["Heat the oven to 200 C. Pat the chickpeas very dry with a clean towel.","Toss chickpeas with olive oil and spread in a single layer on a baking tray.","Roast for 25 to 30 minutes, shaking the tray halfway through, until crisp.","Toss the warm chickpeas with cumin, paprika and salt. Cool slightly before eating."]
      },
      {
        "key":"v1/avocado-tomato-toast", "title":"Avocado and Tomato Toast", "category":"Snacks", "minutes":10,
        "description":"Crisp toast topped with lemony avocado and fresh tomato for a quick, colorful snack.",
        "image":"/img/quickbite.svg",
        "ingredients":["2 slices of bread","1 ripe avocado","1 small tomato, diced","1 teaspoon lemon juice","1 teaspoon olive oil","Salt and black pepper to taste"],
        "steps":["Toast the bread until crisp.","Mash the avocado with lemon juice, salt and pepper.","Spread avocado over the toast and top with diced tomato.","Drizzle with olive oil and serve immediately."]
      }
    ]$recipes$::JSONB)
    LOOP
        IF EXISTS (SELECT 1 FROM public._food_recipe_seed WHERE seed_key = item->>'key') THEN
            CONTINUE;
        END IF;

        INSERT INTO recipe (user_id, category_id, title, description, is_paid, price, prep_time)
        VALUES (author_id, (SELECT id FROM category WHERE category_name = item->>'category'),
            item->>'title', item->>'description', FALSE, 0, (item->>'minutes')::INTEGER)
        RETURNING id INTO new_recipe_id;

        INSERT INTO recipe_ingredient (recipe_id, ingredient)
        SELECT new_recipe_id, value FROM jsonb_array_elements_text(item->'ingredients');

        INSERT INTO recipe_step (recipe_id, step_no, step_description)
        SELECT new_recipe_id, ordinal::INTEGER, value
        FROM jsonb_array_elements_text(item->'steps') WITH ORDINALITY AS step(value, ordinal);

        INSERT INTO recipe_image (recipe_id, img_url, is_thumbnail)
        VALUES (new_recipe_id, item->>'image', TRUE);

        INSERT INTO public._food_recipe_seed (seed_key, recipe_id)
        VALUES (item->>'key', new_recipe_id);
    END LOOP;
END;
$seed$;
COMMIT;
