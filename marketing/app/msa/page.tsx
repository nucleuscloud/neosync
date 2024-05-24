import { Metadata } from 'next';

export const metadata: Metadata = {
  metadataBase: new URL('https://assets.nucleuscloud.com/'),
  title: 'Neosync | MSA',
  description: 'Neosync master services agreement and terms for customers.',
  openGraph: {
    title: 'Neosync',
    description: 'Neosync Master Services Agreement and terms for customers.',
    url: 'https://www.neosync.dev',
    siteName: 'Neosync',
    images: [
      {
        url: '/neosync/marketingsite/mainOGHero.svg',
        width: 1200,
        height: 630,
        alt: 'mainOG',
      },
    ],
    locale: 'en_US',
    type: 'website',
  },
  alternates: {
    canonical: 'https://www.neosync.dev/msa',
  },
};

export default function MSA() {
  return (
    <div className="flex justify-center mx-60 py-40 flex-col gap-4">
      <h1>MASTER SUBSCRIPTION AGREEMENT</h1>
      <p>
        This Nucleus Cloud Corp Master Subscription Agreement (“MSA”) is
        effective as of the effective date of an applicable signed order form
        (such form an “Order Form” and such date the “Effective Date”) and is by
        and between Nucleus Cloud Corp Inc., a Delaware corporation with a place
        of business at 81 Frank Norris St. San Francisco, CA 94109, (“Nucleus
        Cloud Corp”), and the customer set forth on the Order Form (“Customer”)
        (each a “Party” and together the “Parties”). In the event of any
        inconsistency or conflict between the terms of the MSA and the terms of
        any Order Form, the terms of the Order Form control.
      </p>
      <h1>1. Services.</h1>
      <p>
        “Services” means the product(s) and service(s) that are ordered by
        Customer from Nucleus Cloud Corp online or through an Order Form
        referencing this MSA, whether on a trial or paid basis, and to which
        Nucleus Cloud Corp thereby provides access to Customer. Services exclude
        any products or services provided by third parties, even if Customer has
        connected those products or services to the Services. Subject to the
        terms and conditions of this MSA, Nucleus Cloud Corp will make the
        Services available during the Term as set forth in an Order Form.
      </p>
      <h1>2. Fees and Payment. </h1>
      <h2>2.1. Fees. </h2>
      <p>
        Customer will pay the fees specified in the Order Form (the “Fees”).
      </p>
      <h2>2.2. Payment; Taxes </h2>
      <p>
        Nucleus Cloud Corp will invoice Customer for Fees, either within the
        Services or directly, within thirty (30) days of the Effective Date.
        Customer will pay all invoiced Fees net forty-five (45) days from the
        date of the invoice. Any late payments will accrue a 5% interest on the
        monthly payment.. Fees do not include local, state, or federal taxes or
        duties of any kind and any such taxes will be assumed and paid by
        Customer, except for taxes on Nucleus Cloud Corp based on Nucleus Cloud
        Corp’s income or receipts.
      </p>
      <h1>3. Term and Termination. </h1>
      <h2>3.1. Term. </h2>
      <p>
        This MSA commences on the Effective Date and will remain in effect
        through the Initial Term and all Renewal Terms, as specified in the
        Order Form, unless otherwise terminated in accordance with this Section
        (the Initial Term and all Renewal Terms collectively the “Term”).
      </p>
      <h2>3.2. Termination for Cause.</h2>
      <p>
        Either party may terminate this Agreement, effective immediately upon
        written notice to the other party, in the event of a material breach by
        the other party. A material breach shall include, but not be limited to,
        any of the following: Failure to perform services in accordance with the
        terms and conditions of this Agreement; Breach of any confidentiality or
        data protection obligations under this Agreement; Use of the other
        party&apos; intellectual property without permission; Failure to pay any
        amounts due and payable under this Agreement within thirty (30) days
        after receipt of written notice of such non-payment; A change in control
        of the other party or a transfer of substantially all of the assets or
        business of the other party, unless such transfer is to a successor that
        assumes all obligations of this Agreement. In the event of a termination
        for cause, the terminating party shall be entitled to all remedies
        available at law or in equity, including, but not limited to, the right
        to recover damages and to seek injunctive relief.
      </p>
      <h2>3.3. Cancellation. </h2>
      <p>
        Either party may cancel this Agreement upon written notice to the other
        party in the event of a material breach of this Agreement by the other
        party that remains uncured for a period of 45 days following written
        notice of such breach. In addition, either party may cancel this
        Agreement without cause upon 60 days&APOS;prior written notice to the
        other party. Upon cancellation of this Agreement, the parties shall
        immediately cease all further performance of their obligations under
        this Agreement, except for any obligations that survive termination as
        set forth in this Agreement. If the Client cancels this Agreement
        without cause, the Provider shall be entitled to receive payment for all
        services rendered up to the effective date of cancellation, as well as
        any other amounts payable under this Agreement that are due and owing.
        If the Provider cancels this Agreement without cause, the Client shall
        be entitled to a refund of any amounts paid in advance for services that
        have not yet been rendered as of the effective date of cancellation.
        This clause shall not relieve either party from any liabilities or
        obligations that accrued prior to the effective date of cancellation.
      </p>
      <h2>3.4. Effect of Termination and Survival.</h2>
      <p>
        Upon termination or cancellation of an Order Form or this MSA, the
        following provisions shall survive and continue in full force and
        effect: Obligations to Pay: Any obligation of the parties to pay or
        reimburse the other party for services rendered or expenses incurred
        prior to the effective date of termination or expiration shall survive
        such termination or expiration. Confidentiality: The parties&apos;
        obligations regarding confidentiality and non-disclosure shall survive
        such termination or expiration and continue in accordance with the terms
        of this Agreement. Intellectual Property: Any provisions regarding
        intellectual property rights, ownership, and licensing shall survive
        such termination or expiration and continue in accordance with the terms
        of this Agreement. Dispute Resolution: Any provisions regarding dispute
        resolution and governing law shall survive such termination or
        expiration and continue in accordance with the terms of this Agreement.
        Effect of Termination: Upon termination or expiration of this Agreement,
        each party shall immediately return to the other party all property,
        data, and confidential information that belong to the other party. No
        Further Obligations: Except as expressly provided in this clause or
        elsewhere in this Agreement, the parties shall have no further
        obligations or liabilities to each other upon termination or expiration
        of this Agreement. Rights and Remedies: Termination or expiration of
        this Agreement shall not affect any rights or remedies that have accrued
        to either party as of the date of termination or expiration. Survival:
        The provisions of this clause and any other provisions of this Agreement
        that by their nature are intended to survive termination or expiration
        shall survive termination or expiration of this Agreement. Termination
        or expiration of this Agreement shall not relieve the parties of any
        obligations or liabilities incurred prior to such termination or
        expiration. Any provisions of this Agreement that by their nature are
        intended to survive termination or expiration shall survive and continue
        to be binding and enforceable.
      </p>
      <h1>4. License and Use of the Services. </h1>
      <h2>4.1. License.</h2>
      <p>
        Nucleus Cloud Corp hereby grants Customer a non-exclusive,
        non-transferable, worldwide license to use the Neosync Cloud Platform
        (“Licensed Intellectual Property”) solely for the purpose of receiving
        the services provided under this Agreement. Ownership of Licensed
        Intellectual Property: Nucleus Cloud Corp retains all right, title, and
        interest in and to the Licensed Intellectual Property. The Customer
        shall not acquire any rights in the Licensed Intellectual Property
        except as expressly provided in this Agreement. Customer Restrictions:
        The Customer shall not copy, modify, distribute, sell, sublicense, or
        otherwise transfer the Licensed Intellectual Property, in whole or in
        part, except as expressly permitted in this Agreement. The Customer
        shall not use the Licensed Intellectual Property for any purpose other
        than as expressly provided in this Agreement. Term and Termination: This
        license shall commence on the Effective Date and shall continue until
        the termination of this Agreement. Either party may terminate this
        license at any time upon written notice of no less than 60 days to the
        other party in the event of a material breach of this Agreement by the
        other party. Effect of Termination: Upon termination of this license,
        the Customer shall immediately cease all use of the Licensed
        Intellectual Property and shall promptly return or destroy all copies of
        the Licensed Intellectual Property in its possession. No Implied
        Licenses: Except as expressly provided in this Agreement, nothing in
        this Agreement shall be construed as granting any license or rights, by
        implication, estoppel, or otherwise, to any intellectual property of the
        other party. Governing Law and Dispute Resolution: This Agreement shall
        be governed by and construed in accordance with the laws of the state of
        Delaware. Any dispute arising out of or in connection with this
        Agreement shall be resolved in accordance with the dispute resolution
        provisions set forth in this Agreement.
      </p>
      <h2>4.2. Authorized Users.</h2>
      <p>
        Customer may designate and provide access to its (or its corporate
        affiliates’) employees, independent contractors, or other agents to an
        account on the Services as authorized users (each an “Authorized User”)
        up to the number of “seats” set forth in the Order Form (unlimited if
        not specified in the Order Form). Each account may be used only by a
        single, individual Authorized User, and Customer may be charged for
        additional seats (if applicable), or Nucleus Cloud Corp may terminate
        the MSA for cause, if this requirement is circumvented. Customer is
        responsible for all use and misuse of the Services by Authorized User
        accounts and for adherence to this MSA by any Authorized Users, and
        references to Customer herein will be deemed to apply to Authorized
        Users as necessary and applicable. Customer agrees to promptly notify
        Nucleus Cloud Corp of any unauthorized access or use of which Customer
        becomes aware.{' '}
      </p>
      <h2>4.3. Prohibited Uses.</h2>
      <p>
        Customer and Authorized Users will not: (a) “frame,” distribute, resell,
        or permit access to the Services by any third party other than for its
        intended purposes; (b) use the Services other than in compliance with
        applicable federal, state, and local laws; (c) interfere with the
        Services or disrupt any other user’s access to the Subscription Service;
        (d) reverse engineer, attempt to gain unauthorized access to the
        Service, attempt to discover the underlying source code or structure of,
        or otherwise copy or attempt to copy the Services; (e) knowingly
        transfer to the Services any content or data that is defamatory,
        harassing, discriminatory, infringing of third party intellectual
        property rights, or unlawful; (f) transfer to the Services or otherwise
        use on the Services any routine, device, code, exploit, or other
        undisclosed feature that is designed to delete, disable, deactivate,
        interfere with or otherwise harm any software, program, data, device,
        system or service, or which is intended to provide unauthorized access
        or to produce unauthorized modifications; or (g) use any robot, spider,
        data scraping, or extraction tool or similar mechanism with respect to
        the Services.
      </p>
      <h1>5. Confidentiality. </h1>
      <p>
        As used herein, the “Confidential Information” of a Party (the
        “Disclosing Party”) means all financial, technical, or business
        information of the Disclosing Party that the Disclosing Party designates
        as confidential at the time of disclosure to the other Party (the
        “Receiving Party”) or that the Receiving Party reasonably should
        understand to be confidential based on the nature of the information or
        the circumstances surrounding its disclosure. For the sake of clarity,
        the Parties acknowledge that Confidential Information includes the terms
        and conditions of this MSA. Except as expressly permitted in this MSA,
        the Receiving Party will not disclose, duplicate, publish, transfer or
        otherwise make available Confidential Information of the Disclosing
        Party in any form to any person or entity without the Disclosing Party’s
        prior written consent. The Receiving Party will not use the Disclosing
        Party’s Confidential Information except to perform its obligations under
        this MSA, such obligations including, in the case of Nucleus Cloud Corp,
        to provide the Services. Notwithstanding the foregoing, the Receiving
        Party may disclose Confidential Information to the extent required by
        law, provided that the Receiving Party: (a) gives the Disclosing Party
        prior written notice of such disclosure so as to afford the Disclosing
        Party a reasonable opportunity to appear, object, and obtain a
        protective order or other appropriate relief regarding such disclosure
        (if such notice is not prohibited by applicable law); (b) uses diligent
        efforts to limit disclosure and to obtain confidential treatment or a
        protective order; and (c) allows the Disclosing Party to participate in
        the proceeding. Further, Confidential Information does not include any
        information that: (i) is or becomes generally known to the public
        without the Receiving Party&APOS;s breach of any obligation owed to the
        Disclosing Party; (ii) was independently developed by the Receiving
        Party without the Receiving Party&APOS;s breach of any obligation owed
        to the Disclosing Party; or (iii) is received from a third party who
        obtained such Confidential Information without any third party&APOS;s
        breach of any obligation owed to the Disclosing Party.{' '}
      </p>
      <h1>6. Data</h1>
      <h2>6.1 Data Practices.</h2>
      <p>
        {' '}
        Definitions. “Service Data” means a subset of Confidential Information
        comprised of electronic data, text, messages, communications, or other
        materials submitted to and stored within the Services by Customer in
        connection with use of the Services. Service Data may include, without
        limitation, any information relating to an identified or identifiable
        natural person (‘data subject’) where an identifiable natural person is
        one who can be identified, directly or indirectly, in particular by
        reference to an identifier such as name, an identification number,
        location data, an online identifier or to one or more factors specific
        to their physical, physiological, mental, economic, cultural or social
        identity of that natural person (such information, “Personal Data”).
        Service Data does not include metrics and information regarding
        Customer’s use of the Services, including information about how
        Authorized Users use the Services (such information, “Usage Data”).{' '}
      </p>
      <h2>6.2. Ownership.</h2>
      <p>
        Customer will continue to retain its ownership rights to all Service
        Data processed under the terms of this MSA and Nucleus Cloud Corp will
        own all Usage Data.
      </p>
      <h2>6.3. Nucleus Cloud Corp’s Use of Data..</h2>
      <p>
        Nucleus Cloud Corp will not use any personally identificable data
        collected for any other purposes than agreed upon purposes with the
        customer.
      </p>
      <h3>6.3.1. Operating the Services</h3>
      <p>
        Nucleus Cloud Corp may receive, collect, store and/or process Service
        Data based on Nucleus Cloud Corp’s legitimate interest in operating the
        Services. For example, Nucleus Cloud Corp may collect Personal Data
        (such as name, phone number, or credit card information) through the
        account activation process. Nucleus Cloud Corp may also use Service Data
        in an anonymized manner, such as conversion to numerical value, for the
        training of the machine learning models to support certain features and
        functionality within the Services.
      </p>
      <h3>6.3.2. Communications. </h3>
      <p>
        Nucleus Cloud Corp may communicate with Customer or Authorized Users (i)
        to send product information and promotional offers or (i) about the
        Services generally. If Customer or an Authorized User does not want to
        receive such communications, Customer may email
        support@nucleuscloud.com. Customer and necessary Authorized Users will
        always receive transactional messages that are required for Nucleus
        Cloud Corp to provide the Services (such as billing notices and product
        usage notifications).{' '}
      </p>
      <h3>6.3.3. Improving the Services. </h3>
      <p>
        Nucleus Cloud Corp may collect, and may engage third-party analytics
        providers to collect Usage Data to develop new features, improve
        existing features, or inform sales and marketing strategies based on
        Nucleus Cloud Corp’s legitimate interest in improving the Services. Any
        such third-party analytics providers will not share or otherwise
        disclose Usage Data, although Nucleus Cloud Corp may make Usage Data
        publicly available from time to time.{' '}
      </p>
      <h3>6.3.4. Connecting to Third-Party Services. </h3>
      <p>
        Customer may wish to connect third-party services to the Services (e.g.,
        connecting Nucleus Cloud Corp to Customer’s single-sign-on service to
        verify 2FA status of Customer’s employees). When Customer uses a
        third-party service to connect with Nucleus Cloud Corp, logs into the
        Services through a third-party authentication service, or otherwise
        provides Nucleus Cloud Corp with access to information from a
        third-party service, Nucleus Cloud Corp may obtain other information,
        including Personal Data, from those third parties and combine that
        Service or Usage Data based on Nucleus Cloud Corp’s legitimate interest
        in providing Customer with functionality that supports the Services. Any
        access that Nucleus Cloud Corp may receive to such information from a
        third-party service is always in accordance with the features and
        functionality, particularly as to authorization, of that service. By
        authorizing Nucleus Cloud Corp to connect with a third-party service,
        Customer authorizes Nucleus Cloud Corp to access and store any
        information provided to Nucleus Cloud Corp by that third-party service,
        and to use and disclose that information in accordance with this MSA.
      </p>
      <h3>6.3.5. Third-Party Service Providers. </h3>
      <p>
        Customer agrees that Nucleus Cloud Corp may provide Service Data and
        Personal Data to authorized third-party service providers, only to the
        extent necessary to provide, secure, or improve the Services. Any such
        third-party service providers will only be given access to Service Data
        and Personal Data as is reasonably necessary to provide the Services and
        will be subject to (a) confidentiality obligations which are
        commercially reasonable and substantially consistent with the standards
        described in this MSA; and (b) their agreement to comply with the data
        transfer restrictions applicable to Personal Data as set forth below.
        6.4. Service Data Safeguards. (i) Nucleus Cloud Corp will not sell,
        rent, or lease Service Data to any third party, and will not share
        Service Data with third parties, except as permitted by this MSA and to
        provide, secure, and support the Services. (ii) Nucleus Cloud Corp will
        maintain commercially reasonable (particularly for a company of Nucleus
        Cloud Corp’s size and revenue) appropriate administrative, physical, and
        technical safeguards for protection of the security, confidentiality,
        and integrity of Service Data.{' '}
      </p>
      <h1>7. Privacy Practices. </h1>
      <h2>7.1. Privacy Policy.</h2>
      <p>
        Nucleus Cloud Corp operates the Services and, as applicable, handles
        Personal Data, pursuant to the privacy policy available at
        nucleuscloud.com/privacy-policy agree that Customer determines the
        purpose and means of processing such Personal Data, and Nucleus Cloud
        Corp processes such information on behalf of Customer.
      </p>
      <h2>7.2. Hosting and Processing.</h2>
      <p>
        {' '}
        Unless otherwise specifically agreed to by Nucleus Cloud Corp, Service
        Data may be hosted by the Nucleus Cloud Corp, or its respective
        authorized third-party service providers, in the United States or other
        locations around the world. In providing the Services, Nucleus Cloud
        Corp will engage entities to process Service Data, including and without
        limitation, any Personal Data within Service Data pursuant to this MSA,
        within the United States and in other countries and territories.{' '}
      </p>
      <h2>7.3. Sub-Processors. </h2>
      <h3>
        Customer acknowledges and agrees that Nucleus Cloud Corp may use
        third-party data processors engaged by Nucleus Cloud Corp who receive
        Service Data from Nucleus Cloud Corp for processing on behalf of
        Customer and in accordance with Customer’s instructions (as communicated
        by Nucleus Cloud Corp) and the terms of its written subcontract (the
        “Sub-Processors”). Such Sub-Processors may access Service Data to
        provide, secure, and improve the Services. Nucleus Cloud Corp will be
        responsible for the acts and omissions of 7.1. Customer as Controller.
        To the extent Service Data constitutes Personal Data, the Parties
        Sub-Processors to the same extent that Nucleus Cloud Corp would be
        responsible if Nucleus Cloud Corp was performing the services directly
        under the terms of this MSA. The names and locations of all current
        Sub-Processors used for the processing of Personal Data under this MSA,
        if any, are set forth in the Privacy Policy.{' '}
      </h3>
      <h1>8. Intellectual Property Rights. </h1>
      <p>
        Each Party will retain all rights, title and interest in any patents,
        inventions, copyrights, trademarks, domain names, trade secrets,
        know-how and any other intellectual property and/or proprietary rights
        (“Intellectual Property Rights”), and Nucleus Cloud Corp in particular
        will exclusively retain such rights in the Services and all components
        of or used to provide the Services. Customer hereby provides Nucleus
        Cloud Corp a fully paid-up, royalty-free, worldwide, transferable,
        sub-licensable (through multiple layers), assignable, irrevocable and
        perpetual license to implement, use, modify, commercially exploit,
        incorporate into the Services or otherwise use any suggestions,
        enhancement requests, recommendations or other feedback Nucleus Cloud
        Corp receives from Customer, Customer’s agents or representatives,
        Authorized Users, or other third parties acting on Customer’s behalf;
        and Nucleus Cloud Corp also reserves the right to seek intellectual
        property protection for any features, functionality or components that
        may be based on or that were initiated by such suggestions, enhancement
        requests, recommendations or other feedback.{' '}
      </p>
      <h1>9. Representations, Warranties, and Disclaimers. </h1>
      <h2>9.1. Authority.</h2>
      <p>
        {' '}
        Each Party represents that it has validly entered into this MSA and has
        the legal power to do so.{' '}
      </p>
      <h2>9.2. Warranties. </h2>
      <p>
        Nucleus Cloud Corp warrants that during an applicable Term: Authority:
        It has the full right, power, and authority to enter into and perform
        its obligations under this Agreement. Compliance with Laws: It will
        comply with all applicable laws, regulations, and industry standards in
        the performance of its obligations under this Agreement. Services: The
        services provided under this Agreement shall be performed in a
        professional and workmanlike manner, in accordance with generally
        accepted industry standards. No Infringement: The services provided by
        the Provider shall not infringe any third-party intellectual property
        rights, and the Provider has the necessary rights to use and license any
        materials or intellectual property that are provided or used in
        connection with the services. No Conflicts: The performance of its
        obligations under this Agreement shall not violate any agreement,
        obligation, or duty to which it is bound. Ownership: Any deliverables
        provided by the Provider under this Agreement shall be original works of
        authorship, and the Provider shall have all necessary rights, title, and
        interest in and to such deliverables. Disclaimer of Other Warranties:
        Except as expressly set forth in this Agreement, neither party makes any
        other warranties, express or implied, with respect to the services
        provided under this Agreement, and each party expressly disclaims all
        other warranties, including without limitation any implied warranties of
        merchantability or fitness for a particular purpose. Each party
        acknowledges that the other party is relying on the foregoing warranties
        in entering into this Agreement.The warranties set forth in this section
        shall survive termination of this Agreement. Customer’s exclusive
        remedies are those described in Section 3 (Term and Termination) herein.
      </p>
      <h2>9.3. Disclaimers</h2>
      <p>
        EXCEPT AS SPECIFICALLY SET FORTH IN THIS SECTION AND ANY APPLICABLE
        SERVICE LEVEL AGREEMENT, THE SERVICES, INCLUDING ALL SERVER AND NETWORK
        COMPONENTS, ARE PROVIDED ON AN “AS IS” AND “AS AVAILABLE” BASIS, WITHOUT
        ANY WARRANTIES OF ANY KIND TO THE FULLEST EXTENT PERMITTED BY LAW, AND
        Nucleus Cloud Corp EXPRESSLY DISCLAIMS ANY AND ALL WARRANTIES, WHETHER
        EXPRESS OR IMPLIED, INCLUDING, BUT NOT LIMITED TO, ANY IMPLIED
        WARRANTIES OF MERCHANTABILITY, TITLE, FITNESS FOR A PARTICULAR PURPOSE,
        AND NON-INFRINGEMENT. CUSTOMER ACKNOWLEDGES THAT Nucleus Cloud Corp DOES
        NOT WARRANT THAT THE SERVICES WILL BE UNINTERRUPTED, TIMELY, SECURE,
        ERROR FREE, OR FREE FROM VIRUSES OR OTHER MALICIOUS SOFTWARE, AND NO
        INFORMATION OR ADVICE OBTAINED BY CUSTOMER FROM Nucleus Cloud Corp OR
        THROUGH THE SERVICES SHALL CREATE ANY WARRANTY NOT EXPRESSLY STATED IN
        THIS MSA. THE PARTIES ADDITIONALLY AGREE THAT Nucleus Cloud Corp WILL
        HAVE NO LIABILITY OR RESPONSIBILITY FOR CLIENT’S VARIOUS COMPLIANCE
        PROGRAMS, AND THAT THE SERVICES, TO THE EXTENT APPLICABLE, ARE ONLY
        TOOLS FOR ASSISTING CLIENT IN MEETING THE VARIOUS COMPLIANCE OBLIGATIONS
        FOR WHICH IT SOLELY IS RESPONSIBLE.{' '}
      </p>
      <h1>10. Indemnification. </h1>
      <h2>10.1. Indemnification by Nucleus Cloud Corp. </h2>
      <p>
        Nucleus Cloud Corp will indemnify and hold Customer harmless from and
        against any third party claim against Customer alleging that Customer’s
        use of a Service as permitted by this MSA infringes or misappropriates a
        third party’s valid patent, copyright, trademark, or trade secret (an
        “IP Claim”). Nucleus Cloud Corp will, at its expense, defend such IP
        Claim and pay damages finally awarded against Customer in connection
        therewith, including the reasonable fees and expenses of the attorneys
        engaged by Nucleus Cloud Corp for such defense, provided that (a)
        Customer promptly notifies Nucleus Cloud Corp of the threat or notice of
        such IP Claim; (b) Nucleus Cloud Corp will have the sole and exclusive
        control and authority to select defense attorneys, and defend and/or
        settle any such IP Claim (however, Nucleus Cloud Corp will not settle or
        compromise any claim that results in liability or admission of any
        liability by Customer without prior written consent); and (c) Customer
        fully cooperates with Nucleus Cloud Corp in connection therewith. If use
        of a Service by Customer has become, or, in Nucleus Cloud Corp’s
        opinion, is likely to become, the subject of any such IP Claim, Nucleus
        Cloud Corp may, at its option and expense, (i) procure for Customer the
        right to continue using the Service(s) as set forth hereunder; (ii)
        replace or modify a Service to make it non-infringing; or (iii) if
        options (i) or (ii) are not commercially reasonable or practicable as
        determined by Nucleus Cloud Corp, terminate Customer’s subscription to
        the Service(s) and repay, on a pro-rata basis, any Fees previously paid
        to Nucleus Cloud Corp for the corresponding unused portion of the Term
        for such Service(s). Nucleus Cloud Corp will have no liability or
        obligation under this Section with respect to any IP Claim if such claim
        is caused in whole or in part by (x) Nucleus Cloud Corp’s compliance
        with designs, data, instructions, or specifications provided by
        Customer; (y) modification of the Service(s) by anyone other than
        Nucleus Cloud Corp or use of the Service(s) in violation of (i) this
        MSA, (ii) written instructions provided by Nucleus Cloud Corp, or (iii)
        the product features of the Service(s); or (z) the combination,
        operation or use of the Service(s) with other hardware or software where
        a Service would not by itself be infringing. The provisions of this
        Section state the sole, exclusive, and entire liability of Nucleus Cloud
        Corp to Customer and constitute Customer’s sole remedy with respect to
        an IP Claim brought by reason of access to or use of a Service by
        Customer, Customer’s agents, or Authorized Users.
      </p>
      <h2>10.2. Indemnification by Customer. </h2>
      <p>
        Customer will indemnify and hold Nucleus Cloud Corp harmless against any
        third party claim (a) arising from or related to use of a Service by
        Customer, Customer’s agents, or Authorized Users in breach of this MSA;
        or (b) alleging that Customer’s Service Data infringes or
        misappropriates a third party’s valid patent, copyright, trademark, or
        trade secret; provided (i) Nucleus Cloud Corp promptly notifies Customer
        of the threat or notice of such claim; (ii) Customer will have the sole
        and exclusive control and authority to select defense attorneys, and
        defend and/or settle any such claim (however, Customer will not settle
        or compromise any claim that results in liability or admission of any
        liability by Nucleus Cloud Corp without prior written consent); and
        (iii) Nucleus Cloud Corp fully cooperates in connection therewith.
      </p>
      <h1>11. LIMITATION OF LIABILITY. </h1>
      <p>
        UNDER NO CIRCUMSTANCES AND UNDER NO LEGAL THEORY (WHETHER IN CONTRACT,
        TORT, NEGLIGENCE OR OTHERWISE) WILL EITHER PARTY TO THIS MSA, OR THEIR
        AFFILIATES, OFFICERS, DIRECTORS, EMPLOYEES, AGENTS, SERVICE PROVIDERS,
        SUPPLIERS OR LICENSORS BE LIABLE TO THE OTHER PARTY OR ANY AFFILIATE FOR
        ANY LOST PROFITS, LOST SALES OR BUSINESS, LOST DATA (BEING DATA LOST IN
        THE COURSE OF TRANSMISSION VIA CUSTOMER’S SYSTEMS OR OVER THE INTERNET
        THROUGH NO FAULT OF Nucleus Cloud Corp), BUSINESS INTERRUPTION, LOSS OF
        GOODWILL, COSTS OF COVER OR REPLACEMENT, OR FOR ANY TYPE OF INDIRECT,
        INCIDENTAL, SPECIAL, EXEMPLARY, CONSEQUENTIAL, OR PUNITIVE LOSS OR
        DAMAGES, OR ANY OTHER INDIRECT LOSS OR DAMAGES INCURRED BY THE OTHER
        PARTY OR ANY AFFILIATE IN CONNECTION WITH THIS MSA OR THE SERVICES
        REGARDLESS OF WHETHER SUCH PARTY HAS BEEN ADVISED OF THE POSSIBILITY OF
        OR COULD HAVE FORESEEN SUCH DAMAGES. NOTWITHSTANDING ANYTHING TO THE
        CONTRARY IN THIS MSA, AND EXCLUDING THE PARTIES’ INDEMNIFICATION
        OBLIGATIONS HEREUNDER, EITHER PARTY’S AGGREGATE LIABILITY TO THE OTHER
        ARISING OUT OF THIS MSA OR THE SERVICES WILL IN NO EVENT EXCEED
        $100,000. CUSTOMER ACKNOWLEDGES AND AGREES THAT THE ESSENTIAL PURPOSE OF
        THIS SECTION AND THE PARTIES INDEMNIFICATION OBLIGATIONS IS TO ALLOCATE
        THE RISKS UNDER THIS MSA BETWEEN THE PARTIES AND LIMIT POTENTIAL
        LIABILITY GIVEN THE FEES, WHICH WOULD HAVE BEEN SUBSTANTIALLY HIGHER IF
        Nucleus Cloud Corp WERE TO ASSUME ANY FURTHER LIABILITY OTHER THAN AS
        SET FORTH HEREIN. Nucleus Cloud Corp HAS RELIED ON THESE LIMITATIONS IN
        DETERMINING WHETHER TO PROVIDE CUSTOMER WITH THE RIGHTS TO ACCESS AND
        USE THE SERVICES PROVIDED FOR IN THIS MSA.{' '}
      </p>
      <h1> 12. Miscellaneous. </h1>
      <p>
        12.1. Entire Agreement. This MSA and the applicable Order Form(s)
        constitute the entire agreement, and supersedes all prior agreements,
        between Nucleus Cloud Corp and Customer regarding the subject matter
        hereof.
      </p>
      <h2>12.2. Assignment.</h2>
      <p>
        {' '}
        Either Party may, without the consent of the other Party, assign this
        MSA to any affiliate or in connection with any merger, change of
        control, or the sale of all or substantially all of such Party’s assets
        provided that (1) the other Party is provided prior notice of such
        assignment and (2) any such successor agrees to fulfill its obligations
        pursuant to this MSA. Subject to the foregoing restrictions, this MSA
        will be fully binding upon, inure to the benefit of and be enforceable
        by the Parties and their respective successors and assigns.{' '}
      </p>
      <h2>12.3. Severability.</h2>{' '}
      <p>
        {' '}
        If any provision in this MSA is held by a court of competent
        jurisdiction to be unenforceable, such provision will be modified by the
        court and interpreted so as to best accomplish the original provision to
        the fullest extent permitted by law, and the remaining provisions of
        this MSA will remain in effect.{' '}
      </p>
      <h2>12.4. Relationship of the Parties.</h2>{' '}
      <p>
        The Parties are independent contractors. This MSA does not create a
        partnership, franchise, joint venture, agency, fiduciary, or employment
        relationship between the Parties.
      </p>
      <h2>12.5. Notices.</h2>{' '}
      <p>
        All notices provided by Nucleus Cloud Corp to Customer under this MSA
        may be delivered in writing (a) by nationally recognized overnight
        delivery service (“Courier”) or U.S. mail to the contact mailing address
        provided by Customer on the Order Form; or (b) electronic mail to the
        electronic mail address provided for Customer’s account owner. Customer
        must give notice to Nucleus Cloud Corp in writing by Courier or U.S.
        mail to 81 Frank Norris St. apt 604 San Francisco, California. All
        notices shall be deemed to have been given immediately upon delivery by
        electronic mail; or, if otherwise delivered upon the earlier of receipt
        or two (2) business days after being deposited in the mail or with a
        Courier as permitted above.{' '}
      </p>
      <h2>12.6. Governing Law, Jurisdiction, Venue.</h2>{' '}
      <p>
        {' '}
        This MSA will be governed by the laws of the State of California,
        without reference to conflict of laws principles. Any disputes under
        this MSA shall be resolved in a court of general jurisdiction in San
        Francisco County, California. Customer hereby expressly agrees to submit
        to the exclusive personal jurisdiction and venue of such courts for the
        purpose of resolving any dispute relating to this MSA or access to or
        use of the Services by Customer, its agents, or Authorized Users.{' '}
      </p>
      <h2>12.7. Export Compliance.</h2>{' '}
      <p>
        The Services and other software or components of the Services that
        Nucleus Cloud Corp may provide or make available to Customer are subject
        to U.S. export control and economic sanctions laws as administered and
        enforced by the Office of Foreign Assets and Control of the United
        States Department of Treasury. Customer agrees to comply with all such
        laws and regulations as they relate to access to and use of the
        Services. Customer will not access or use the Services if Customer or
        any Authorized Users are located in any jurisdiction in which the
        provision of the Services, software, or other components is prohibited
        under U.S. or other applicable laws or regulations (a “Prohibited
        Jurisdiction”) and Customer will not provide access to the Services to
        any government, entity, or individual located in any Prohibited
        Jurisdiction. Customer represents and warrants that (a) it is not named
        on any U.S. government list of persons or entities prohibited from
        receiving U.S. exports, or transacting with any U.S. person; (b) it is
        not a national of, or a company registered in, any Prohibited
        Jurisdiction; (c) it will not permit any individuals under its control
        to access or use the Services in violation of any U.S. or other
        applicable export embargoes, prohibitions or restrictions; and (d) it
        will comply with all applicable laws regarding the transmission of
        technical data exported from the United States and the countries in
        which it and Authorized Users are located.{' '}
      </p>
      <h2>12.8. Anti-Corruption.</h2>
      <p>
        {' '}
        Customer agrees that it has not received or been offered any illegal or
        improper bribe, kickback, payment, gift, or thing of value from any of
        Nucleus Cloud Corp’s employees or agents in connection with this MSA.
        Reasonable gifts and entertainment provided in the ordinary course of
        business do not violate the above restriction. If Customer learns of any
        violation of the above restriction, Customer will use reasonable efforts
        to promptly give notice to Nucleus Cloud Corp. `
      </p>
      <h2>12.9. Publicity and Marketing.</h2>{' '}
      <p>
        Nucleus Cloud Corp may use Customer’s name, logo, and trademarks solely
        to identify Customer as a client of Nucleus Cloud Corp on Nucleus Cloud
        Corp’s website and other marketing materials and in accordance with
        Customer’s trademark usage guidelines, if Customer provides same to
        Nucleus Cloud Corp. Nucleus Cloud Corp may share aggregated and/or
        anonymized information regarding use of the Services with third parties
        for marketing purposes to develop and promote Services. Nucleus Cloud
        Corp never will disclose aggregated and/or anonymized information to a
        third party in a manner that would identify Customer as the source of
        the information or Authorized Users or others personally.{' '}
      </p>
      <h2>12.10. Amendments.</h2>{' '}
      <p>
        Nucleus Cloud Corp may amend this MSA from time to time, in which case
        the new MSA will supersede prior versions. Nucleus Cloud Corp will
        notify Customer not less than ten (10) days prior to the effective date
        of any such amendment and Customer’s continued use of the Services
        following the effective date of any such amendment may be relied upon by
        Nucleus Cloud Corp as consent to any such amendment. Nucleus Cloud
        Corp’s failure to enforce at any time any provision of this MSA does not
        constitute a waiver of that provision or of any other provision of this
        MSA.
      </p>
    </div>
  );
}
