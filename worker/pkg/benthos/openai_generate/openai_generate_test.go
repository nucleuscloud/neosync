package openaigenerate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	content = `content,id,created_at,sender_email,sender_name,receiver_email,receiver_name,subject
"Dear John,\n\nI hope this email finds you well. Following up on our previous discussion, could you please provide the latest updates on PO #12345? We need to finalize the order by the end of this week.\n\nBest regards,\nAnna","1","2023-09-21 15:23:45 +00:00","anna.smith@newcompany.com","Anna Smith","john.doe@oldcompany.com","John Doe","PO #12345 Updates"
"Hi Anna,\n\nSure, I will get back to you by tomorrow with the necessary details and updates. Thanks for your patience.\n\nRegards,\nJohn","2","2023-09-21 17:12:01 +00:00","john.doe@oldcompany.com","John Doe","anna.smith@newcompany.com","Anna Smith","Re: PO #12345 Updates"
"Dear Maria,\n\nCould you please expedite the processing of order #67890? We are on a tight schedule and need the materials delivered sooner than expected.\n\nThank you,\nJames","3","2023-09-20 10:05:34 +00:00","james.lee@anothercompany.com","James Lee","maria.gonzalez@oldcompany.com","Maria Gonzalez","Order #67890 Expedited Processing"
"Hello James,\n\nI will prioritize your order and coordinate with the logistics team to ensure expedited delivery. I will update you shortly.\n\nBest,\nMaria","4","2023-09-20 12:22:45 +00:00","maria.gonzalez@oldcompany.com","Maria Gonzalez","james.lee@anothercompany.com","James Lee","Re: Order #67890 Expedited Processing"
"Dear Alice,\n\nPlease find attached the latest list of approved vendors for your perusal. Let me know if you have any questions.\n\nRegards,\nRobert","5","2023-09-19 14:45:09 +00:00","robert.brown@newcompany.com","Robert Brown","alice.taylor@oldercompany.com","Alice Taylor","Approved Vendor List"
"Hi Robert,\n\nThank you for the list. I will go through it and get back to you with any queries. Much appreciated.\n\nBest,\nAlice","6","2023-09-19 16:18:33 +00:00","alice.taylor@oldercompany.com","Alice Taylor","robert.brown@newcompany.com","Robert Brown","Re: Approved Vendor List"
"Dear Sarah,\n\nI would like to schedule a meeting to discuss potential savings through bulk purchasing of the raw materials. Please let me know your availability.\n\nThank you,\nDavid","7","2023-09-18 09:37:56 +00:00","david.wilson@incompany.com","David Wilson","sarah.jones@oldcompany.com","Sarah Jones","Meeting Request: Bulk Purchasing Savings"
"Hello David,\n\nI am available for a meeting this Friday at 2 PM. Does that time work for you?\n\nRegards,\nSarah","8","2023-09-18 11:45:27 +00:00","sarah.jones@oldcompany.com","Sarah Jones","david.wilson@incompany.com","David Wilson","Re: Meeting Request: Bulk Purchasing Savings"
"Dear Emma,\n\nKindly review the attached proposal for the new blend of chemical solutions. Any feedback will be highly appreciated.\n\nThanks,\nMichael","9","2023-09-17 13:56:40 +00:00","michael.johnson@newcompany.com","Michael Johnson","emma.watson@oldercompany.com","Emma Watson","Proposal Review"
"Hi Michael,\n\nI have received the proposal and will review it with my team. We will get back to you with our comments early next week.\n\nRegards,\nEmma","10","2023-09-17 15:22:59 +00:00","emma.watson@oldercompany.com","Emma Watson","michael.johnson@newcompany.com","Michael Johnson","Re: Proposal Review"
"Dear Andrew,\n\nI wanted to touch base regarding the pricing negotiations. Are there any updates? Our deadline for finalizing the prices is fast approaching.\n\nBest,\nSophia","11","2023-09-16 08:45:00 +00:00","sophia.martin@newcompany.com","Sophia Martin","andrew.clark@oldercompany.com","Andrew Clark","Pricing Negotiations Update"
"Hi Sophia,\n\nI am currently awaiting approval from our senior management. I will update you by the end of business tomorrow.\n\nRegards,\nAndrew","12","2023-09-16 10:15:43 +00:00","andrew.clark@oldercompany.com","Andrew Clark","sophia.martin@newcompany.com","Sophia Martin","Re: Pricing Negotiations Update"
"Dear Olivia,\n\nLooking forward to the meeting next week. Please see the attached agenda for your reference.\n\nThank you,\nEthan","13","2023-09-15 13:14:22 +00:00","ethan.green@newcompany.com","Ethan Green","olivia.james@oldcompany.com","Olivia James","Meeting Agenda"
"Hi Ethan,\n\nThanks for the agenda. I will review it and come prepared for our meeting. See you next week.\n\nBest,\nOlivia","14","2023-09-15 15:21:39 +00:00","olivia.james@oldcompany.com","Olivia James","ethan.green@newcompany.com","Ethan Green","Re: Meeting Agenda"
"Dear Daniel,\n\nPlease find the revised contract terms attached. Let me know if you need any further amendments.\n\nBest,\nLucas","15","2023-09-14 14:30:45 +00:00","lucas.smith@anothercompany.com","Lucas Smith","daniel.bennett@oldercompany.com","Daniel Bennett","Revised Contract Terms"
"Hi Lucas,\n\nThank you for the revised terms. I will review and discuss them with our legal team. I will get back to you by the end of the week.\n\nRegards,\nDaniel","16","2023-09-14 16:45:09 +00:00","daniel.bennett@oldercompany.com","Daniel Bennett","lucas.smith@anothercompany.com","Lucas Smith","Re: Revised Contract Terms"
"Dear Emily,\n\nWe have received your PO #55566. The order is being processed and we will update you upon shipment.\n\nBest,\nSamuel","17","2023-09-13 09:25:35 +00:00","samuel.walker@newcompany.com","Samuel Walker","emily.roberts@oldcompany.com","Emily Roberts","PO #55566 Received"
"Hello Samuel,\n\nThanks for the update. Please keep me informed about the shipment status.\n\nBest regards,\nEmily","18","2023-09-13 11:37:50 +00:00","emily.roberts@oldcompany.com","Emily Roberts","samuel.walker@newcompany.com","Samuel Walker","Re: PO #55566 Received"
"Dear Christina,\n\nI am contacting you regarding our upcoming partnership. Could you please confirm the details of our agreement?\n\nThank you,\nBrian","19","2023-09-12 08:45:56 +00:00","brian.hall@somecompany.com","Brian Hall","christina.collins@oldcompany.com","Christina Collins","Partnership Agreement Details"
"Hi Brian,\n\nYes, I have attached the confirmation details for your review. Let me know if you have any questions.\n\nBest,\nChristina","20","2023-09-12 10:25:43 +00:00","christina.collins@oldcompany.com","Christina Collins","brian.hall@somecompany.com","Brian Hall","Re: Partnership Agreement Details"
"Dear Jake,\n\nThank you for providing the draft MOU. Can we schedule a call to discuss some modifications?\n\nBest,\nLisa","21","2023-09-11 14:23:45 +00:00","lisa.carter@anothercompany.com","Lisa Carter","jake.harris@oldcompany.com","Jake Harris","MOU Discussion"
"Hi Lisa,\n\nOf course, I'm available for a call tomorrow afternoon. Does 3 PM work for you?\n\nRegards,\nJake","22","2023-09-11 16:20:33 +00:00","jake.harris@oldcompany.com","Jake Harris","lisa.carter@anothercompany.com","Lisa Carter","Re: MOU Discussion"
"Danny,\n\nWe need to finalize the contract by EOD. Find attached the latest version.\n\nBest,\nAaron","23","2023-09-10 13:45:34 +00:00","aaron.moore@newcompany.com","Aaron Moore","danny.morris@oldcompany.com","Danny Morris","Contract Finalization"
"Aaron,\n\nReviewing the contract now. Will revert shortly.\n\n- Danny","24","2023-09-10 15:17:26 +00:00","danny.morris@oldcompany.com","Danny Morris","aaron.moore@newcompany.com","Aaron Moore","Re: Contract Finalization"
"Dear Sophie,\n\nI'm reaching out regarding the quarterly audit. Could you provide the necessary documents?\n\nThank you,\nBen","25","2023-09-09 11:38:45 +00:00","ben.jackson@newcompany.com","Ben Jackson","sophie.mitchell@oldcompany.com","Sophie Mitchell","Quarterly Audit Documents"
"Hi Ben,\n\nI have attached the required documents for the audit. Let me know if you need anything else.\n\nBest,\nSophie","26","2023-09-09 13:57:34 +00:00","sophie.mitchell@oldcompany.com","Sophie Mitchell","ben.jackson@newcompany.com","Ben Jackson","Re: Quarterly Audit Documents"
"Dear Amy,\n\nWe have resolved the issue with your PO #99988. The order is now ready for dispatch.\n\nBest,\nElijah","27","2023-09-08 12:10:23 +00:00","elijah.white@newcompany.com","Elijah White","amy.wright@oldercompany.com","Amy Wright","PO #99988 Issue Resolved"
"Hi Elijah,\n\nThank you for the swift resolution. Please provide the tracking information once dispatched.\n\nRegards,\nAmy","28","2023-09-08 14:22:59 +00:00","amy.wright@oldercompany.com","Amy Wright","elijah.white@newcompany.com","Elijah White","Re: PO #99988 Issue Resolved"
"Dear Jason,\n\nWe need your help with the integration of the new system. Can you allocate time for this project?\n\nThank you,\nHenry","29","2023-09-07 09:36:45 +00:00","henry.clark@anothercompany.com","Henry Clark","jason.young@oldcompany.com","Jason Young","Integration Project"
"Hi Henry,\n\nI will check my availability and revert back by tomorrow. Thanks for reaching out.\n\nBest,\nJason","30","2023-09-07 11:15:34 +00:00","jason.young@oldcompany.com","Jason Young","henry.clark@anothercompany.com","Henry Clark","Re: Integration Project"
"Dear Karen,\n\nFollowing up on your inquiry about our new product line. Please find attached the specifications.\n\nBest,\nEmma","31","2023-09-06 14:05:22 +00:00","emma.wood@newcompany.com","Emma Wood","karen.brown@oldercompany.com","Karen Brown","New Product Line Specifications"
"Hi Emma,\n\nThanks for sharing the specs. Reviewing them, and will reach out if we have any questions.\n\nBest,\nKaren","32","2023-09-06 16:18:33 +00:00","karen.brown@oldercompany.com","Karen Brown","emma.wood@newcompany.com","Emma Wood","Re: New Product Line Specifications"
"Dear Charles,\n\nWe're excited about the upcoming collaboration. Could you confirm the kickoff meeting details?\n\nRegards,\nNancy","33","2023-09-05 13:45:56 +00:00","nancy.cooper@somecompany.com","Nancy Cooper","charles.green@oldcompany.com","Charles Green","Collaboration Kickoff Meeting"
"Hi Nancy,\n\nThe meeting is scheduled for September 10th at 10 AM. See you then!\n\nBest,\nCharles","34","2023-09-05 15:57:43 +00:00","charles.green@oldcompany.com","Charles Green","nancy.cooper@somecompany.com","Nancy Cooper","Re: Collaboration Kickoff Meeting"
"Dear Susan,\n\nI hope you are doing well. I'm writing regarding our annual review. Could we meet to discuss the results?\n\nBest,\nGeorge","35","2023-09-04 10:25:34 +00:00","george.mason@newcompany.com","George Mason","susan.hill@oldcompany.com","Susan Hill","Annual Review Discussion"
"Hello George,\n\nSure, I am available on Wednesday afternoon. Does that work for you?\n\nRegards,\nSusan","36","2023-09-04 12:15:45 +00:00","susan.hill@oldcompany.com","Susan Hill","george.mason@newcompany.com","George Mason","Re: Annual Review Discussion"
"Dear Alex,\n\nI am writing to remind you about the product demonstration scheduled for next week. Please confirm your availability.\n\nThanks,\nChris","37","2023-09-03 16:38:27 +00:00","chris.baker@anothercompany.com","Chris Baker","alex.scott@oldcompany.com","Alex Scott","Product Demonstration Confirmation"
"Hi Chris,\n\nI am available as scheduled. Looking forward to it.\n\nBest,\nAlex","38","2023-09-03 18:57:11 +00:00","alex.scott@oldcompany.com","Alex Scott","chris.baker@anothercompany.com","Chris Baker","Re: Product Demonstration Confirmation"
"Hello Rachel,\n\nWe've reviewed your proposal and have a few questions. Can we set up a meeting to discuss this further?\n\nThank you,\nBill","39","2023-09-02 09:45:34 +00:00","bill.michelson@newcompany.com","Bill Michelson","rachel.morgan@oldercompany.com","Rachel Morgan","Proposal Discussion"
"Hi Bill,\n\nAbsolutely, let's schedule a call for this Friday. Does 11 AM work for you?\n\nRegards,\nRachel","40","2023-09-02 11:38:47 +00:00","rachel.morgan@oldercompany.com","Rachel Morgan","bill.michelson@newcompany.com","Bill Michelson","Re: Proposal Discussion"
"Dear Laura,\n\nI wanted to follow up on the delivery date for order #33377. Could we reschedule the delivery for an earlier date?\n\nThank you,\nSteve","41","2023-09-01 13:25:56 +00:00","steve.keith@newcompany.com","Steve Keith","laura.adams@oldercompany.com","Laura Adams","Delivery Date Reschedule"
"Hi Steve,\n\nI will check with our logistics team and get back to you by tomorrow.\n\nBest,\nLaura","42","2023-09-01 15:22:34 +00:00","laura.adams@oldercompany.com","Laura Adams","steve.keith@newcompany.com","Steve Keith","Re: Delivery Date Reschedule"
"Dear Mark,\n\nWould you be able to provide a status update on the pending invoice #45678? We're closing our Q3 books.\n\nThanks,\nHannah","43","2023-08-31 08:16:43 +00:00","hannah.jones@newcompany.com","Hannah Jones","mark.patterson@oldercompany.com","Mark Patterson","Invoice #45678 Status Update"
"Hi Hannah,\n\nThe payment has been processed and will reflect in your account shortly.\n\nRegards,\nMark","44","2023-08-31 10:45:32 +00:00","mark.patterson@oldercompany.com","Mark Patterson","hannah.jones@newcompany.com","Hannah Jones","Re: Invoice #45678 Status Update"
"Dear Jessica,\n\nPlease see the attached reports for the monthly performance review. Let me know if you need any clarifications.\n\nBest,\nDavid","45","2023-08-30 14:36:47 +00:00","david.anderson@anothercompany.com","David Anderson","jessica.davis@oldercompany.com","Jessica Davis","Monthly Performance Review"
"Hi David,\n\nThank you for the reports. I will review them and reach out if I have any questions.\n\nRegards,\nJessica","46","2023-08-30 16:20:31 +00:00","jessica.davis@oldercompany.com","Jessica Davis","david.anderson@anothercompany.com","David Anderson","Re: Monthly Performance Review"
"Dear Megan,\n\nI'm writing to request a detailed statement of account for the past six months for reconciliation purposes.\n\nThanks,\nHarold","47","2023-08-29 12:45:09 +00:00","harold.evans@newcompany.com","Harold Evans","megan.lee@oldcompany.com","Megan Lee","Account Statement Request"
"Hi Harold,\n\nI have compiled the requested statement. Please find it attached.\n\nBest,\nMegan","48","2023-08-29 14:37:45 +00:00","megan.lee@oldcompany.com","Megan Lee","harold.evans@newcompany.com","Harold Evans","Re: Account Statement Request"
"Dear Laura,\n\nI hope this email finds you well. Could you confirm the specs for the equipment we've leased? It’s for our upcoming inspection.\n\nThank you,\nNancy","49","2023-08-28 11:15:34 +00:00","nancy.robinson@anothercompany.com","Nancy Robinson","laura.martinez@oldercompany.com","Laura Martinez","Equipment Specs Confirmation"
"Hi Nancy,\n\nPlease find the confirmed specs attached. Let me know if you have any further questions.\n\nBest,\nLaura","50","2023-08`
)

func Test_getCsvRecordsFromInput_Success(t *testing.T) {
	type testcase struct {
		input    string
		expected [][]string
	}

	testcases := []testcase{
		{input: content, expected: [][]string{{"content", "id", "created_at", "sender_email", "sender_name", "receiver_email", "receiver_name", "subject"}, {"Dear John,\\n\\nI hope this email finds you well. Following up on our previous discussion, could you please provide the latest updates on PO #12345? We need to finalize the order by the end of this week.\\n\\nBest regards,\\nAnna", "1", "2023-09-21 15:23:45 +00:00", "anna.smith@newcompany.com", "Anna Smith", "john.doe@oldcompany.com", "John Doe", "PO #12345 Updates"}, {"Hi Anna,\\n\\nSure, I will get back to you by tomorrow with the necessary details and updates. Thanks for your patience.\\n\\nRegards,\\nJohn", "2", "2023-09-21 17:12:01 +00:00", "john.doe@oldcompany.com", "John Doe", "anna.smith@newcompany.com", "Anna Smith", "Re: PO #12345 Updates"}, {"Dear Maria,\\n\\nCould you please expedite the processing of order #67890? We are on a tight schedule and need the materials delivered sooner than expected.\\n\\nThank you,\\nJames", "3", "2023-09-20 10:05:34 +00:00", "james.lee@anothercompany.com", "James Lee", "maria.gonzalez@oldcompany.com", "Maria Gonzalez", "Order #67890 Expedited Processing"}, {"Hello James,\\n\\nI will prioritize your order and coordinate with the logistics team to ensure expedited delivery. I will update you shortly.\\n\\nBest,\\nMaria", "4", "2023-09-20 12:22:45 +00:00", "maria.gonzalez@oldcompany.com", "Maria Gonzalez", "james.lee@anothercompany.com", "James Lee", "Re: Order #67890 Expedited Processing"}, {"Dear Alice,\\n\\nPlease find attached the latest list of approved vendors for your perusal. Let me know if you have any questions.\\n\\nRegards,\\nRobert", "5", "2023-09-19 14:45:09 +00:00", "robert.brown@newcompany.com", "Robert Brown", "alice.taylor@oldercompany.com", "Alice Taylor", "Approved Vendor List"}, {"Hi Robert,\\n\\nThank you for the list. I will go through it and get back to you with any queries. Much appreciated.\\n\\nBest,\\nAlice", "6", "2023-09-19 16:18:33 +00:00", "alice.taylor@oldercompany.com", "Alice Taylor", "robert.brown@newcompany.com", "Robert Brown", "Re: Approved Vendor List"}, {"Dear Sarah,\\n\\nI would like to schedule a meeting to discuss potential savings through bulk purchasing of the raw materials. Please let me know your availability.\\n\\nThank you,\\nDavid", "7", "2023-09-18 09:37:56 +00:00", "david.wilson@incompany.com", "David Wilson", "sarah.jones@oldcompany.com", "Sarah Jones", "Meeting Request: Bulk Purchasing Savings"}, {"Hello David,\\n\\nI am available for a meeting this Friday at 2 PM. Does that time work for you?\\n\\nRegards,\\nSarah", "8", "2023-09-18 11:45:27 +00:00", "sarah.jones@oldcompany.com", "Sarah Jones", "david.wilson@incompany.com", "David Wilson", "Re: Meeting Request: Bulk Purchasing Savings"}, {"Dear Emma,\\n\\nKindly review the attached proposal for the new blend of chemical solutions. Any feedback will be highly appreciated.\\n\\nThanks,\\nMichael", "9", "2023-09-17 13:56:40 +00:00", "michael.johnson@newcompany.com", "Michael Johnson", "emma.watson@oldercompany.com", "Emma Watson", "Proposal Review"}, {"Hi Michael,\\n\\nI have received the proposal and will review it with my team. We will get back to you with our comments early next week.\\n\\nRegards,\\nEmma", "10", "2023-09-17 15:22:59 +00:00", "emma.watson@oldercompany.com", "Emma Watson", "michael.johnson@newcompany.com", "Michael Johnson", "Re: Proposal Review"}, {"Dear Andrew,\\n\\nI wanted to touch base regarding the pricing negotiations. Are there any updates? Our deadline for finalizing the prices is fast approaching.\\n\\nBest,\\nSophia", "11", "2023-09-16 08:45:00 +00:00", "sophia.martin@newcompany.com", "Sophia Martin", "andrew.clark@oldercompany.com", "Andrew Clark", "Pricing Negotiations Update"}, {"Hi Sophia,\\n\\nI am currently awaiting approval from our senior management. I will update you by the end of business tomorrow.\\n\\nRegards,\\nAndrew", "12", "2023-09-16 10:15:43 +00:00", "andrew.clark@oldercompany.com", "Andrew Clark", "sophia.martin@newcompany.com", "Sophia Martin", "Re: Pricing Negotiations Update"}, {"Dear Olivia,\\n\\nLooking forward to the meeting next week. Please see the attached agenda for your reference.\\n\\nThank you,\\nEthan", "13", "2023-09-15 13:14:22 +00:00", "ethan.green@newcompany.com", "Ethan Green", "olivia.james@oldcompany.com", "Olivia James", "Meeting Agenda"}, {"Hi Ethan,\\n\\nThanks for the agenda. I will review it and come prepared for our meeting. See you next week.\\n\\nBest,\\nOlivia", "14", "2023-09-15 15:21:39 +00:00", "olivia.james@oldcompany.com", "Olivia James", "ethan.green@newcompany.com", "Ethan Green", "Re: Meeting Agenda"}, {"Dear Daniel,\\n\\nPlease find the revised contract terms attached. Let me know if you need any further amendments.\\n\\nBest,\\nLucas", "15", "2023-09-14 14:30:45 +00:00", "lucas.smith@anothercompany.com", "Lucas Smith", "daniel.bennett@oldercompany.com", "Daniel Bennett", "Revised Contract Terms"}, {"Hi Lucas,\\n\\nThank you for the revised terms. I will review and discuss them with our legal team. I will get back to you by the end of the week.\\n\\nRegards,\\nDaniel", "16", "2023-09-14 16:45:09 +00:00", "daniel.bennett@oldercompany.com", "Daniel Bennett", "lucas.smith@anothercompany.com", "Lucas Smith", "Re: Revised Contract Terms"}, {"Dear Emily,\\n\\nWe have received your PO #55566. The order is being processed and we will update you upon shipment.\\n\\nBest,\\nSamuel", "17", "2023-09-13 09:25:35 +00:00", "samuel.walker@newcompany.com", "Samuel Walker", "emily.roberts@oldcompany.com", "Emily Roberts", "PO #55566 Received"}, {"Hello Samuel,\\n\\nThanks for the update. Please keep me informed about the shipment status.\\n\\nBest regards,\\nEmily", "18", "2023-09-13 11:37:50 +00:00", "emily.roberts@oldcompany.com", "Emily Roberts", "samuel.walker@newcompany.com", "Samuel Walker", "Re: PO #55566 Received"}, {"Dear Christina,\\n\\nI am contacting you regarding our upcoming partnership. Could you please confirm the details of our agreement?\\n\\nThank you,\\nBrian", "19", "2023-09-12 08:45:56 +00:00", "brian.hall@somecompany.com", "Brian Hall", "christina.collins@oldcompany.com", "Christina Collins", "Partnership Agreement Details"}, {"Hi Brian,\\n\\nYes, I have attached the confirmation details for your review. Let me know if you have any questions.\\n\\nBest,\\nChristina", "20", "2023-09-12 10:25:43 +00:00", "christina.collins@oldcompany.com", "Christina Collins", "brian.hall@somecompany.com", "Brian Hall", "Re: Partnership Agreement Details"}, {"Dear Jake,\\n\\nThank you for providing the draft MOU. Can we schedule a call to discuss some modifications?\\n\\nBest,\\nLisa", "21", "2023-09-11 14:23:45 +00:00", "lisa.carter@anothercompany.com", "Lisa Carter", "jake.harris@oldcompany.com", "Jake Harris", "MOU Discussion"}, {"Hi Lisa,\\n\\nOf course, I'm available for a call tomorrow afternoon. Does 3 PM work for you?\\n\\nRegards,\\nJake", "22", "2023-09-11 16:20:33 +00:00", "jake.harris@oldcompany.com", "Jake Harris", "lisa.carter@anothercompany.com", "Lisa Carter", "Re: MOU Discussion"}, {"Danny,\\n\\nWe need to finalize the contract by EOD. Find attached the latest version.\\n\\nBest,\\nAaron", "23", "2023-09-10 13:45:34 +00:00", "aaron.moore@newcompany.com", "Aaron Moore", "danny.morris@oldcompany.com", "Danny Morris", "Contract Finalization"}, {"Aaron,\\n\\nReviewing the contract now. Will revert shortly.\\n\\n- Danny", "24", "2023-09-10 15:17:26 +00:00", "danny.morris@oldcompany.com", "Danny Morris", "aaron.moore@newcompany.com", "Aaron Moore", "Re: Contract Finalization"}, {"Dear Sophie,\\n\\nI'm reaching out regarding the quarterly audit. Could you provide the necessary documents?\\n\\nThank you,\\nBen", "25", "2023-09-09 11:38:45 +00:00", "ben.jackson@newcompany.com", "Ben Jackson", "sophie.mitchell@oldcompany.com", "Sophie Mitchell", "Quarterly Audit Documents"}, {"Hi Ben,\\n\\nI have attached the required documents for the audit. Let me know if you need anything else.\\n\\nBest,\\nSophie", "26", "2023-09-09 13:57:34 +00:00", "sophie.mitchell@oldcompany.com", "Sophie Mitchell", "ben.jackson@newcompany.com", "Ben Jackson", "Re: Quarterly Audit Documents"}, {"Dear Amy,\\n\\nWe have resolved the issue with your PO #99988. The order is now ready for dispatch.\\n\\nBest,\\nElijah", "27", "2023-09-08 12:10:23 +00:00", "elijah.white@newcompany.com", "Elijah White", "amy.wright@oldercompany.com", "Amy Wright", "PO #99988 Issue Resolved"}, {"Hi Elijah,\\n\\nThank you for the swift resolution. Please provide the tracking information once dispatched.\\n\\nRegards,\\nAmy", "28", "2023-09-08 14:22:59 +00:00", "amy.wright@oldercompany.com", "Amy Wright", "elijah.white@newcompany.com", "Elijah White", "Re: PO #99988 Issue Resolved"}, {"Dear Jason,\\n\\nWe need your help with the integration of the new system. Can you allocate time for this project?\\n\\nThank you,\\nHenry", "29", "2023-09-07 09:36:45 +00:00", "henry.clark@anothercompany.com", "Henry Clark", "jason.young@oldcompany.com", "Jason Young", "Integration Project"}, {"Hi Henry,\\n\\nI will check my availability and revert back by tomorrow. Thanks for reaching out.\\n\\nBest,\\nJason", "30", "2023-09-07 11:15:34 +00:00", "jason.young@oldcompany.com", "Jason Young", "henry.clark@anothercompany.com", "Henry Clark", "Re: Integration Project"}, {"Dear Karen,\\n\\nFollowing up on your inquiry about our new product line. Please find attached the specifications.\\n\\nBest,\\nEmma", "31", "2023-09-06 14:05:22 +00:00", "emma.wood@newcompany.com", "Emma Wood", "karen.brown@oldercompany.com", "Karen Brown", "New Product Line Specifications"}, {"Hi Emma,\\n\\nThanks for sharing the specs. Reviewing them, and will reach out if we have any questions.\\n\\nBest,\\nKaren", "32", "2023-09-06 16:18:33 +00:00", "karen.brown@oldercompany.com", "Karen Brown", "emma.wood@newcompany.com", "Emma Wood", "Re: New Product Line Specifications"}, {"Dear Charles,\\n\\nWe're excited about the upcoming collaboration. Could you confirm the kickoff meeting details?\\n\\nRegards,\\nNancy", "33", "2023-09-05 13:45:56 +00:00", "nancy.cooper@somecompany.com", "Nancy Cooper", "charles.green@oldcompany.com", "Charles Green", "Collaboration Kickoff Meeting"}, {"Hi Nancy,\\n\\nThe meeting is scheduled for September 10th at 10 AM. See you then!\\n\\nBest,\\nCharles", "34", "2023-09-05 15:57:43 +00:00", "charles.green@oldcompany.com", "Charles Green", "nancy.cooper@somecompany.com", "Nancy Cooper", "Re: Collaboration Kickoff Meeting"}, {"Dear Susan,\\n\\nI hope you are doing well. I'm writing regarding our annual review. Could we meet to discuss the results?\\n\\nBest,\\nGeorge", "35", "2023-09-04 10:25:34 +00:00", "george.mason@newcompany.com", "George Mason", "susan.hill@oldcompany.com", "Susan Hill", "Annual Review Discussion"}, {"Hello George,\\n\\nSure, I am available on Wednesday afternoon. Does that work for you?\\n\\nRegards,\\nSusan", "36", "2023-09-04 12:15:45 +00:00", "susan.hill@oldcompany.com", "Susan Hill", "george.mason@newcompany.com", "George Mason", "Re: Annual Review Discussion"}, {"Dear Alex,\\n\\nI am writing to remind you about the product demonstration scheduled for next week. Please confirm your availability.\\n\\nThanks,\\nChris", "37", "2023-09-03 16:38:27 +00:00", "chris.baker@anothercompany.com", "Chris Baker", "alex.scott@oldcompany.com", "Alex Scott", "Product Demonstration Confirmation"}, {"Hi Chris,\\n\\nI am available as scheduled. Looking forward to it.\\n\\nBest,\\nAlex", "38", "2023-09-03 18:57:11 +00:00", "alex.scott@oldcompany.com", "Alex Scott", "chris.baker@anothercompany.com", "Chris Baker", "Re: Product Demonstration Confirmation"}, {"Hello Rachel,\\n\\nWe've reviewed your proposal and have a few questions. Can we set up a meeting to discuss this further?\\n\\nThank you,\\nBill", "39", "2023-09-02 09:45:34 +00:00", "bill.michelson@newcompany.com", "Bill Michelson", "rachel.morgan@oldercompany.com", "Rachel Morgan", "Proposal Discussion"}, {"Hi Bill,\\n\\nAbsolutely, let's schedule a call for this Friday. Does 11 AM work for you?\\n\\nRegards,\\nRachel", "40", "2023-09-02 11:38:47 +00:00", "rachel.morgan@oldercompany.com", "Rachel Morgan", "bill.michelson@newcompany.com", "Bill Michelson", "Re: Proposal Discussion"}, {"Dear Laura,\\n\\nI wanted to follow up on the delivery date for order #33377. Could we reschedule the delivery for an earlier date?\\n\\nThank you,\\nSteve", "41", "2023-09-01 13:25:56 +00:00", "steve.keith@newcompany.com", "Steve Keith", "laura.adams@oldercompany.com", "Laura Adams", "Delivery Date Reschedule"}, {"Hi Steve,\\n\\nI will check with our logistics team and get back to you by tomorrow.\\n\\nBest,\\nLaura", "42", "2023-09-01 15:22:34 +00:00", "laura.adams@oldercompany.com", "Laura Adams", "steve.keith@newcompany.com", "Steve Keith", "Re: Delivery Date Reschedule"}, {"Dear Mark,\\n\\nWould you be able to provide a status update on the pending invoice #45678? We're closing our Q3 books.\\n\\nThanks,\\nHannah", "43", "2023-08-31 08:16:43 +00:00", "hannah.jones@newcompany.com", "Hannah Jones", "mark.patterson@oldercompany.com", "Mark Patterson", "Invoice #45678 Status Update"}, {"Hi Hannah,\\n\\nThe payment has been processed and will reflect in your account shortly.\\n\\nRegards,\\nMark", "44", "2023-08-31 10:45:32 +00:00", "mark.patterson@oldercompany.com", "Mark Patterson", "hannah.jones@newcompany.com", "Hannah Jones", "Re: Invoice #45678 Status Update"}, {"Dear Jessica,\\n\\nPlease see the attached reports for the monthly performance review. Let me know if you need any clarifications.\\n\\nBest,\\nDavid", "45", "2023-08-30 14:36:47 +00:00", "david.anderson@anothercompany.com", "David Anderson", "jessica.davis@oldercompany.com", "Jessica Davis", "Monthly Performance Review"}, {"Hi David,\\n\\nThank you for the reports. I will review them and reach out if I have any questions.\\n\\nRegards,\\nJessica", "46", "2023-08-30 16:20:31 +00:00", "jessica.davis@oldercompany.com", "Jessica Davis", "david.anderson@anothercompany.com", "David Anderson", "Re: Monthly Performance Review"}, {"Dear Megan,\\n\\nI'm writing to request a detailed statement of account for the past six months for reconciliation purposes.\\n\\nThanks,\\nHarold", "47", "2023-08-29 12:45:09 +00:00", "harold.evans@newcompany.com", "Harold Evans", "megan.lee@oldcompany.com", "Megan Lee", "Account Statement Request"}, {"Hi Harold,\\n\\nI have compiled the requested statement. Please find it attached.\\n\\nBest,\\nMegan", "48", "2023-08-29 14:37:45 +00:00", "megan.lee@oldcompany.com", "Megan Lee", "harold.evans@newcompany.com", "Harold Evans", "Re: Account Statement Request"}, {"Dear Laura,\\n\\nI hope this email finds you well. Could you confirm the specs for the equipment we've leased? It’s for our upcoming inspection.\\n\\nThank you,\\nNancy", "49", "2023-08-28 11:15:34 +00:00", "nancy.robinson@anothercompany.com", "Nancy Robinson", "laura.martinez@oldercompany.com", "Laura Martinez", "Equipment Specs Confirmation"}}},
		{
			input:    "id,name,email\n1,nick,nick@example.com\n2,nick2,nick2@example.com",
			expected: [][]string{{"id", "name", "email"}, {"1", "nick", "nick@example.com"}, {"2", "nick2", "nick2@example.com"}},
		},
		{
			input:    "```csv\nid,name,email\n1,nick,nick@example.com\n2,nick2,nick2@example.com\n```",
			expected: [][]string{{"id", "name", "email"}, {"1", "nick", "nick@example.com"}, {"2", "nick2", "nick2@example.com"}},
		},
	}

	for _, tc := range testcases {
		t.Run(t.Name(), func(t *testing.T) {
			actual, err := getCsvRecordsFromInput(tc.input, nil)
			require.NoError(t, err)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func Test_convertCsvToStructuredRecord_Success(t *testing.T) {
	type testcase struct {
		record  []string
		headers []string
		types   []string
	}

	testcases := []testcase{
		{
			record:  []string{},
			headers: []string{},
			types:   []string{},
		},
		{
			headers: []string{"c1", "c2", "c3", "c4"},
			types:   []string{"smallint", "integer", "int", "serial"},
			record:  []string{"1", "2", "3", "4"},
		},
		{
			headers: []string{"c1", "c2"},
			types:   []string{"bigint", "bigserial"},
			record:  []string{"1", "2"},
		},
		{
			headers: []string{"c1", "c2"},
			types:   []string{"numeric", "decimal"},
			record:  []string{"1.11", "2.22"},
		},
		{
			headers: []string{"c1"},
			types:   []string{"real"},
			record:  []string{"1.11"},
		},
		{
			headers: []string{"c1"},
			types:   []string{"double precision"},
			record:  []string{"1.11"},
		},
		{
			headers: []string{"c1"},
			types:   []string{"money"},
			record:  []string{"1.11"},
		},
		{
			headers: []string{"c1", "c2", "c3", "c4", "c5"},
			types:   []string{"character varying", "varchar", "character", "char", "text"},
			record:  []string{"1", "2", "3", "4", "5"},
		},
		{
			headers: []string{"c1", "c2", "c3"},
			types:   []string{"date", "timestamp", "timestamp without time zone"},
			record:  []string{"1", "2", "3"},
		},
		{
			headers: []string{"c1"},
			types:   []string{"timestamp with time zone"},
			record:  []string{"1"},
		},
		{
			headers: []string{"c1", "c2"},
			types:   []string{"time", "time without time zone"},
			record:  []string{"1", "2"},
		},
		{
			headers: []string{"c1"},
			types:   []string{"time with time zone"},
			record:  []string{"1"},
		},
		{
			headers: []string{"c1"},
			types:   []string{"interval"},
			record:  []string{"1"},
		},
		{
			headers: []string{"c1"},
			types:   []string{"boolean"},
			record:  []string{"true"},
		},
		{
			headers: []string{"c1"},
			types:   []string{"uuid"},
			record:  []string{"1"},
		},
		{
			headers: []string{"c1", "c2"},
			types:   []string{"json", "jsonb"},
			record:  []string{"1", "2"},
		},
		{
			headers: []string{"c1"},
			types:   []string{"text[]"},
			record:  []string{`["1","2"]`},
		},
		{
			headers: []string{"c1"},
			types:   []string{"unknown custom type"},
			record:  []string{"1"},
		},
	}

	for _, tc := range testcases {
		t.Run("", func(t *testing.T) {
			_, err := convertCsvToStructuredRecord(tc.record, tc.headers, tc.types)
			require.NoError(t, err)
		})
	}
}
