
<!-- SOLVED -->
1. When a user confirms an order the confirmation email (asynchronously) is not to the user’s email address.

<!-- SOLVED -->
2. The go-backend is not connected to the postgres-mcp server yet so the {{base_url}}/chat/prompt endpoint is working but giving incorrect results.

<!-- Steps to handle problem 2 -->
-Collect a catalog of items and products to put into the database to be used as the bench mark.
-Users can order what is in the database and it is delivered if product is not in the database then the application should tell the user that "We do not have %v that the moment. productUserRequested"


3. Should the admin endpoints use the same JWT token as the regular user endpoints!!
4. I am getting a database query error on this {{base_url}}/admin/config (Admin endpoint) my guess is that these details are not in the database they are just in the code.
5. There is a database query error for this endpoint too {{base_url}}/admin/config because these details are not in the database

<!-- SOLVED -->
6. There needs to be a special email like deliveryjaj@gmail.com that sends these details to the users.  `

<!-- SOLVED -->
7. The tokens sent to the users need to be more user friendly not just ( ie 546JSKSN7822....)

<!-- SOLVED -->
8. The signup should have username, email, password instead of email , password

9. The Logo in the email is not showing

<!-- SOLVED -->
10. User to confirm order after requesting for a product.

11. User is able to specify the order in terms of brand of the order just like i would like Ntake Bread not Supa Loaf and if the application only has one brand it can tell the user that only that brand is available.

12. User should be able to only place an order once in a day. If that works then user should also be able to place many orders throughout the day this implies that the program will be adding to their order of that day.When a new day comes new new chat hence new orders

13. Password Reset is not working in the frontend but it works in the backend

<!-- SOLVED -->
14. Implement sessions and cookies for atleast 6 months.