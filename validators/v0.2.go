package validators;

import (
	"github.com/go-playground/validator/v10"
	"github.com/thoas/go-funk"
)

func isCategory_v0_2(fl validator.FieldLevel) bool {
	var supportedCategories = []string{
		"accounting",
		"agile-project-management",
		"applicant-tracking",
		"application-development",
		"appointment-scheduling",
		"backup",
		"billing-and-invoicing",
		"blog",
		"budgeting",
		"business-intelligence",
		"business-process-management",
		"cad",
		"call-center-management",
		"cloud-management",
		"collaboration",
		"communications",
		"compliance-management",
		"contact-management",
		"content-management",
		"crm",
		"customer-service-and-support",
		"data-analytics",
		"data-collection",
		"data-visualization",
		"digital-asset-management",
		"digital-citizenship",
		"document-management",
		"donor-management",
		"e-commerce",
		"e-signature",
		"educational-content",
		"email-management",
		"email-marketing",
		"employee-management",
		"enterprise-project-management",
		"enterprise-social-networking",
		"erp",
		"event-management",
		"facility-management",
		"feedback-and-reviews-management",
		"financial-reporting",
		"fleet-management",
		"fundraising",
		"gamification",
		"geographic-information-systems",
		"grant-management",
		"graphic-design",
		"help-desk",
		"hr",
		"ide",
		"identity-management",
		"instant-messaging",
		"inventory-management",
		"it-asset-management",
		"it-development",
		"it-management",
		"it-security",
		"it-service-management",
		"knowledge-management",
		"learning-management-system",
		"marketing",
		"mind-mapping",
		"mobile-marketing",
		"mobile-payment",
		"network-management",
		"office",
		"online-booking",
		"online-community",
		"payment-gateway",
		"payroll",
		"predictive-analysis",
		"procurement",
		"productivity-suite",
		"project-collaboration",
		"project-management",
		"property-management",
		"real-estate-management",
		"remote-support",
		"resource-management",
		"sales-management",
		"seo",
		"service-desk",
		"social-media-management",
		"survey",
		"talent-management",
		"task-management",
		"taxes-management",
		"test-management",
		"time-management",
		"time-tracking",
		"translation",
		"video-conferencing",
		"video-editing",
		"visitor-management",
		"voip",
		"warehouse-management",
		"web-collaboration",
		"web-conferencing",
		"whistleblowing",
		"website-builder",
		"workflow-management",
	}

	return funk.Contains(supportedCategories, fl.Field().String())
}

func isScope_v0_2(fl validator.FieldLevel) bool {
	var supportedScopes = []string{
		"agriculture",
		"culture",
		"defence",
		"education",
		"emergency-services",
		"employment",
		"energy",
		"environment",
		"finance-and-economic-development",
		"foreign-affairs",
		"government",
		"healthcare",
		"infrastructures",
		"justice",
		"local-authorities",
		"manufacturing",
		"research",
		"science-and-technology",
		"security",
		"society",
		"sport",
		"tourism",
		"transportation",
		"welfare",
	}

	return funk.Contains(supportedScopes, fl.Field().String())
}
