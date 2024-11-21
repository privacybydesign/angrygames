(function ($) {
	"use strict"

	// Mobile Nav toggle
	$('.menu-toggle > a').on('click', function (e) {
		e.preventDefault();
		$('#responsive-nav').toggleClass('active');
	})

	// Fix cart dropdown from closing
	$('.cart-dropdown').on('click', function (e) {
		e.stopPropagation();
	});

	/////////////////////////////////////////

	// Products Slick
	$('.products-slick').each(function () {
		var $this = $(this),
			$nav = $this.attr('data-nav');

		$this.slick({
			slidesToShow: 4,
			slidesToScroll: 1,
			autoplay: true,
			infinite: true,
			speed: 300,
			dots: false,
			arrows: true,
			appendArrows: $nav ? $nav : false,
			responsive: [{
					breakpoint: 991,
					settings: {
						slidesToShow: 2,
						slidesToScroll: 1,
					}
				},
				{
					breakpoint: 480,
					settings: {
						slidesToShow: 1,
						slidesToScroll: 1,
					}
				},
			]
		});
	});

	// Product Main img Slick
	$('#product-main-img').slick({
		infinite: true,
		speed: 300,
		dots: false,
		arrows: true,
		fade: true,
		asNavFor: '#product-imgs',
	});

	// Product imgs Slick
	$('#product-imgs').slick({
		slidesToShow: 3,
		slidesToScroll: 1,
		arrows: true,
		centerMode: true,
		focusOnSelect: true,
		centerPadding: 0,
		vertical: true,
		asNavFor: '#product-main-img',
		responsive: [{
			breakpoint: 991,
			settings: {
				vertical: false,
				arrows: false,
				dots: true,
			}
		}, ]
	});

	// Product img zoom
	var zoomMainProduct = document.getElementById('product-main-img');
	if (zoomMainProduct) {
		$('#product-main-img .product-preview').zoom();
	}

	// Go to age verification page
	$('.add-to-cart-btn').on('click', function (element) {
		sessionStorage.setItem('minage', element.currentTarget.getAttribute('data-minage'));
		sessionStorage.setItem('productname', element.currentTarget.getAttribute('data-productname'));
		sessionStorage.setItem('productid', element.currentTarget.id);
		window.location.href = 'cart.html';
	});

	// If no product is chosen yet, return to the index page.
	if (location.pathname.includes('cart.html') && !sessionStorage.getItem('productid'))
		location.href = 'index.html';

	// Age verification with Yivi
	if ( window.yivi ) {
		console.log("Age verification started");

		let yiviCore = yivi.newWeb({
			element:   '#yivi-web-element',

			session: {
				url: '..',
				start: {
					url: o => `start_session.php?type=` + sessionStorage.getItem('minage') + 'plus',
				},
				result: {
					url: (o, {sessionPtr, sessionToken}) => `${sessionPtr.u.split('/irma')[0]}/session/${sessionToken}/result`,
				}
			},

			translations: {
				'success': 'Gegevens ontvangen.'
			},
		});

		let success = function (data) {
			console.log("Session successful!");
			console.log("Result:", data);
			// Continue to order page if user is 18+
			let attr = daawvalue.toLowerCase();
			if (attr === 'yes' || attr === 'ja') {
				setTimeout(() => {
					$('#phase-agecheck').hide();
					$('#phase-finished').show();
					$('#breadcrumb').show();
					$('html').scrollTop(0);
				}, 1000);
			} else {
				$('#snackbar-content').html('U hebt niet de juiste leeftijd om dit artikel te mogen bestellen.');
				$('#snackbar').addClass('show');
			}
		};

		let error = function (data) {
			console.log("Session failed!");
			console.log("Error data:", data);
			$('#snackbar-content').html('Er is een fout opgetreden bij de leeftijdsverificatie via Yivi.');
			$('#snackbar').addClass('show');
		};

		yiviCore.start().then(success, error);
	}

})(jQuery);
ta.disclosed[0][0].r