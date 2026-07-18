# Yakusoku Ledger features

## Student workflow

1. Enroll a Student organization identity.
2. Enter the agreement details and choose the supporting document.
3. The browser calculates the document's SHA-256 fingerprint locally.
4. Only the fingerprint and agreement metadata are submitted to Fabric.
5. Track the submitted agreement and its university decision in the dashboard.

The source document is never uploaded to Yakusoku Ledger. A later copy can be selected
in the verification panel; its local fingerprint is compared with the immutable value.

## University workflow

University organization members enrolled with `UNIVERSITY_ENROLLMENT_SECRET` see
submitted agreements in the review queue. An
approval or rejection creates a new ledger version rather than replacing history. The
chaincode independently verifies that the reviewer belongs to `UniversityMSP`.

## Dashboard and analytics

The dashboard calculates these values from live ledger records:

- total agreements
- pending approvals
- approved agreement value
- agreements carrying verified document fingerprints

Before authentication, clearly marked preview records demonstrate the product without
pretending that sample data came from Fabric.

## Audit history and notifications

Every agreement row opens its complete Fabric key history, including transaction ID,
timestamp, status, and value. Successful enrollment, submission, review, and document
checks also appear in the activity feed as browser notifications.

## Roles

| Identity | Capabilities |
| --- | --- |
| Student organization member | Submit, browse, verify, and audit agreements |
| University organization member | All read operations plus approve/reject |
| Network administrator | Channel, peer, chaincode installation, and instantiation |

The network administrator secret only grants API lifecycle permissions. Agreement
review authorization is enforced from the transaction creator's Fabric MSP.
