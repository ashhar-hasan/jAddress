## Existing Logic in INDFAS

All the relevant code can be found in `alice/alice/protected/modules/mobapi/controllers/CustomerController.php`.

### actionAddNewAddress

#### DATA

- *HTTP_PARAMS*: `is_exchange`, `create_ticket`
- *SESSION*: `returnRequest`, `exhangeRequest`
- *Forward Serviceability Condition*: `isPrecious`, `isFragile`, `dispatchLoc`

#### LOGIC

- If `is_exchange` is set, get the `exchangeRequest` and set the Forward Serviceability Conditions array.
- Read the address form.
- [Cleanup](#cleanupaddress) the `address1` and `address2` fields.
- If the `postcode` is not empty, check the Reverse Serviceability of the postcode.
- If the form validates, the pincode is reverse serviceable and address type is set,
  - If the address is not default billing address, then set the phone number to `COUNTRY_CODE + phone`.

### cleanUpAddress

- Remove all characters except `A-Z`, `a-z`, `0-9`, any whitespace, `.:-/%+()*[]{}$#!@"';\n\t\r`. The PCRE regex is `'/[^A-Za-z0-9\s+\.\:\-\/%\+\(\)\*\[\]\{\}\$\#\!\@\"\';\n\t\r]/'`.
- If the resulting string is not empty, trim whitespaces and `.:-/%()*&[]{}$#!@"';`.

