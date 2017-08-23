# jAuth-new Proposed Architecture

## Endpoints

- `POST /address`: Create a new address
  ```json
  {
    "Address1": "string",
    "Address2": "string",
    "AddressRegion": "string",
    "AddressType": "string",
    "AlternatePhone": "string",
    "City": "string",
    "Country": "string",
    "FirstName": "string",
    "LastName": "string",
    "Phone": "string",
    "PostCode": "string",
    "RegionName": "string",
    "Sms_opt": "string"
  }
  ```

- `DELETE /address/{id}`: Delete address by id
- `PUT /address/{id}`: Update address by id
  ```json
  {
    "Address1": "string",
    "Address2": "string",
    "AddressRegion": "string",
    "AddressType": "string",
    "AlternatePhone": "string",
    "City": "string",
    "Country": "string",
    "FirstName": "string",
    "LastName": "string",
    "Phone": "string",
    "PostCode": "string",
    "RegionName": "string",
    "Sms_opt": "string"
  }
  ```

- `GET /address/{type}`: Get address by type
- `PUT /address/{type}/{id}`: Set default billing or shipping address
- `GET /address/locality/{pincode}`: Get locality by pincode
  ```json
  {
    "messages": {},
    "metadata": {
      "city": [
        {
          "state": "string",
          "stateId": "string",
          "value": "string"
        }
      ],
      "city_name": "string",
      "id_customer_address_region": "string",
      "locality": [
        null
      ],
      "pincode": "string",
      "state": "string"
    },
    "session-id": "string",
    "success": true
  }
  ```

## Workflow Definition

- Request Validator:
  - Check required request params are present or not
  - Check that correct HTTP verb is used with the corresponding request

### Get Locality:
- Request Validator
- Get Locality:
  - Check if pin code is present and is a number
  - Check if available in cache
  - Retrieve from database if cache miss and set in cache

### Delete Address:
- Request Validator
- Delete Address:
  - Check that the address to be deleted is not the default shipping address or the default billing address
  - Check if address present in cache, delete from cache and database
  - Use a transaction while deleting from database
  - If transaction fails, rollback the transaction

### Update Address, Add Address and Set Address Type:
- Request Validator
- Address Validator:
  - *Id*, *Phone*, *AlternatePhone*, *AddressRegion*, *Country* should be int
  - *FirstName*, *LastName*, *Address1*, *Address2*, *City* should be string
  - *Phone* and *AlternatePhone* should be 10 digit
  - *Postcode* should be int and 6 digits
  - *sms_opt* and *is_office* is a flag and should be either 0 or 1
  - *AddressType* can be **"billing"**, **"shipping"**, **"other"** or **"all"**
  - *Req* can be either **" "** or **"update_type"**

  For update, *HTTP Verb* == PUT, *Req* == "update_type", *Id* != 0, *AddressType* != "".
  Required parameters are *Id* != 0, *FirstName*, *Address1*, *City*, *PostCode*, *AddressRegion*.  
  If *Req* == "update_type" then the *AddressType* is updated else the entire address is updated.  
  For adding new address, *HTTP Verb* == POST and *Id* is missing.  
  Required parameters are *FirstName*, *Address1*, *City*, *PostCode*, *AddressRegion*.

- Data Encryptor:
  - Send a request to the encryption service and use the response for both encryption and decryption.

### List Address:

- Request Validator
- Query Term Validator:
  - Check that *limit* is a valid number
  - If *limit* > *MAX_LIMIT*, then *limit* = *DEFAULT_LIMIT*
  - Check that *offset* is a valid number
  - Check that *AddressType* is not empty and is one of **all**, **billing**, **shipping**, **other**
- List Address:
  - Retrieve address list from cache
  - Retrieve address list from database if cache miss
    - Hit the decryption service to decrypt encrypted fields
  - Filter the retrieved addresses based on *AddressType*

## Quirks

### Deletion

**WHAT**: During deletion, the address is deleted from the cache and then a goroutine is called to delete from database.  
**JUSTIFICATION**: It is very rare that a deletion may fail and it is not critical. Also database code is slower than cache code.  
**DRAWBACK**: This means that if the deletion from database fails for some reason the user will (incorrectly) see that their address has been deleted.

### Creation

**WHAT**: During creation, the address is added to the database and then a goroutine is called to add the newly created address to the cache.  
**JUSTIFICATION**: A temp *Id* will have to be created to be stored in the cache and then the cache will have to be updated with the actual *Id* after the database code runs.  
**DRAWBACK**: None

